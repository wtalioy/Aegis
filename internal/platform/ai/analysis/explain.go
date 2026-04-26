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

	if view, ok := storage.View(event); ok {
		pid = view.PID
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

	actions := generateSuggestedActions(req.EventID, matchedRule)

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

func generateSuggestedActions(eventID string, rule *policy.Rule) []types.Action {
	actions := []types.Action{}

	if rule != nil && rule.IsTesting() {
		actions = append(actions, types.Action{
			Label:    "转正规则",
			ActionID: "promote",
			Params:   types.ActionParams{RuleName: rule.Name},
		})
	}

	actions = append(actions, types.Action{
		Label:    "调查",
		ActionID: "investigate",
		Params:   types.ActionParams{EventID: eventID},
	})

	return actions
}

func relatedEventFromStorage(event *storage.Event) types.RelatedEvent {
	if event == nil {
		return types.RelatedEvent{}
	}

	related := types.RelatedEvent{
		Timestamp: event.Timestamp.UnixMilli(),
	}

	view, ok := storage.View(event)
	if !ok {
		return related
	}

	switch view.Type {
	case events.EventTypeExec:
		related.Type = "exec"
		related.PPID = view.PPID
	case events.EventTypeFileOpen:
		related.Type = "file"
		related.Filename = view.Filename
	case events.EventTypeConnect:
		related.Type = "connect"
		related.Port = view.Port
	}
	related.PID = view.PID
	related.CgroupID = fmt.Sprintf("%d", view.CgroupID)
	related.ProcessName = view.ProcessName
	related.Blocked = view.Blocked

	return related
}
