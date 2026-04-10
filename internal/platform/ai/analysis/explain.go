package analysis

import (
	"context"
	"fmt"
	"strings"

	"aegis/internal/analysis/types"
	"aegis/internal/platform/ai/prompt"
	"aegis/internal/platform/ai/providers"
	"aegis/internal/platform/ai/snapshot"
	"aegis/internal/platform/events"
	"aegis/internal/platform/storage"
	"aegis/internal/policy"
	"aegis/internal/shared/utils"
	"aegis/internal/telemetry/proc"
	"aegis/internal/telemetry/workload"
)

func ExplainEvent(
	ctx context.Context,
	p providers.Provider,
	req *types.ExplainRequest,
	event *storage.Event,
	ruleEngine *policy.Engine,
	store storage.EventStore,
	profileReg *proc.ProfileRegistry,
	workloadReg *workload.Registry,
	processTree *proc.ProcessTree,
	snapshotState *snapshot.SystemState,
) (*types.ExplainResponse, error) {
	if p == nil {
		return nil, fmt.Errorf("AI provider is not available")
	}

	var relatedEvents []types.RelatedEvent
	var pid uint32

	switch ev := event.Data.(type) {
	case *events.ExecEvent:
		pid = ev.Hdr.PID
	case *events.FileOpenEvent:
		pid = ev.Hdr.PID
	case *events.ConnectEvent:
		pid = ev.Hdr.PID
	case map[string]any:
		if v, ok := ev["pid"].(float64); ok {
			pid = uint32(v)
		} else if hdr, ok := ev["header"].(map[string]any); ok {
			if v, ok := hdr["pid"].(float64); ok {
				pid = uint32(v)
			}
		}
	}

	if store != nil && pid != 0 {
		relatedEvents = []types.RelatedEvent{relatedEventFromStorage(event)}
	}

	var profile *proc.ProcessProfile
	if profileReg != nil && pid != 0 {
		profile, _ = profileReg.GetProfile(pid)
	}

	eventDesc := prompt.FormatEventForExplanation(event, profile)
	question := req.Question

	// Build related context
	relatedContext := prompt.BuildExplainContext(event, pid, store, profileReg, workloadReg, processTree, snapshotState)

	fullPrompt := prompt.BuildExplainPrompt(eventDesc, question, relatedContext)
	response, err := p.SingleChat(ctx, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI inference failed: %w", err)
	}

	explanation := response
	rootCause := extractRootCause(response)

	var matchedRule *policy.Rule
	if ruleEngine != nil {
		// Hook for rule matching in future.
		_ = matchedRule
	}

	actions := generateSuggestedActions(event, matchedRule)

	return &types.ExplainResponse{
		Explanation:      explanation,
		RootCause:        rootCause,
		MatchedRule:      matchedRule,
		RelatedEvents:    relatedEvents,
		SuggestedActions: actions,
	}, nil
}

func extractRootCause(response string) string {
	// Simple heuristic: return the first non-empty line as the root cause summary.
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func generateSuggestedActions(event *storage.Event, rule *policy.Rule) []types.Action {
	actions := []types.Action{}

	if rule != nil && rule.IsTesting() {
		actions = append(actions, types.Action{
			Label:    "转正规则",
			ActionID: "promote",
			Params:   map[string]any{"rule_name": rule.Name},
		})
	}

	actions = append(actions, types.Action{
		Label:    "调查",
		ActionID: "investigate",
		Params:   map[string]any{"event_id": fmt.Sprintf("%v", event)},
	})

	return actions
}

func relatedEventFromStorage(event *storage.Event) types.RelatedEvent {
	if event == nil {
		return types.RelatedEvent{}
	}

	related := types.RelatedEvent{
		Timestamp: event.Timestamp.UnixMilli(),
		Details:   map[string]any{},
	}

	switch ev := event.Data.(type) {
	case *events.ExecEvent:
		related.Type = "exec"
		related.PID = ev.Hdr.PID
		related.CgroupID = fmt.Sprintf("%d", ev.Hdr.CgroupID)
		related.ProcessName = utils.ExtractCString(ev.Hdr.Comm[:])
		related.Blocked = ev.Hdr.Blocked == 1
		related.Details["ppid"] = ev.PPID
	case events.ExecEvent:
		related.Type = "exec"
		related.PID = ev.Hdr.PID
		related.CgroupID = fmt.Sprintf("%d", ev.Hdr.CgroupID)
		related.ProcessName = utils.ExtractCString(ev.Hdr.Comm[:])
		related.Blocked = ev.Hdr.Blocked == 1
		related.Details["ppid"] = ev.PPID
	case *events.FileOpenEvent:
		related.Type = "file"
		related.PID = ev.Hdr.PID
		related.CgroupID = fmt.Sprintf("%d", ev.Hdr.CgroupID)
		related.ProcessName = utils.ExtractCString(ev.Hdr.Comm[:])
		related.Blocked = ev.Hdr.Blocked == 1
		related.Details["filename"] = utils.ExtractCString(ev.Filename[:])
	case events.FileOpenEvent:
		related.Type = "file"
		related.PID = ev.Hdr.PID
		related.CgroupID = fmt.Sprintf("%d", ev.Hdr.CgroupID)
		related.ProcessName = utils.ExtractCString(ev.Hdr.Comm[:])
		related.Blocked = ev.Hdr.Blocked == 1
		related.Details["filename"] = utils.ExtractCString(ev.Filename[:])
	case *events.ConnectEvent:
		related.Type = "connect"
		related.PID = ev.Hdr.PID
		related.CgroupID = fmt.Sprintf("%d", ev.Hdr.CgroupID)
		related.ProcessName = utils.ExtractCString(ev.Hdr.Comm[:])
		related.Blocked = ev.Hdr.Blocked == 1
		related.Details["port"] = ev.Port
	case events.ConnectEvent:
		related.Type = "connect"
		related.PID = ev.Hdr.PID
		related.CgroupID = fmt.Sprintf("%d", ev.Hdr.CgroupID)
		related.ProcessName = utils.ExtractCString(ev.Hdr.Comm[:])
		related.Blocked = ev.Hdr.Blocked == 1
		related.Details["port"] = ev.Port
	}

	return related
}
