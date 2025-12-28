package analysis

import (
	"context"
	"fmt"
	"strconv"

	"aegis/pkg/ai/prompt"
	"aegis/pkg/ai/providers"
	"aegis/pkg/ai/snapshot"
	"aegis/pkg/ai/types"
	"aegis/pkg/proc"
	"aegis/pkg/rules"
	"aegis/pkg/workload"
)

func Analyze(
	ctx context.Context,
	p providers.Provider,
	req *types.AnalyzeRequest,
	profileReg *proc.ProfileRegistry,
	workloadReg *workload.Registry,
	ruleEngine *rules.Engine,
	snapshotState *snapshot.SystemState,
) (*types.AnalyzeResponse, error) {
	if p == nil {
		return nil, fmt.Errorf("AI provider is not available")
	}

	var (
		analysisData string
		anomalies    []types.Anomaly
		snapshotCtx  string
	)

	switch req.Type {
	case types.AnalyzeTypeProcess:
		pid, err := strconv.ParseUint(req.ID, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid PID: %w", err)
		}
		profile, ok := profileReg.GetProfile(uint32(pid))
		if !ok {
			return nil, fmt.Errorf("process profile not found for PID %d", pid)
		}
		analysisData = prompt.FormatProcessProfile(profile)
		anomalies = analyzeProcessAnomalies(profile)

	case types.AnalyzeTypeWorkload:
		cgroupID, err := strconv.ParseUint(req.ID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid CgroupID: %w", err)
		}
		w := workloadReg.Get(cgroupID)
		if w == nil {
			return nil, fmt.Errorf("workload not found for CgroupID %d", cgroupID)
		}
		analysisData = prompt.FormatWorkloadMetadata(w)
		anomalies = analyzeWorkloadAnomalies(w)

	case types.AnalyzeTypeRule:
		allRules := ruleEngine.GetRules()
		var rule *rules.Rule
		for i := range allRules {
			if allRules[i].Name == req.ID {
				rule = &allRules[i]
				break
			}
		}
		if rule == nil {
			return nil, fmt.Errorf("rule not found: %s", req.ID)
		}
		analysisData = prompt.FormatRuleForAnalysis(rule, ruleEngine)
		anomalies = analyzeRuleAnomalies(rule, ruleEngine)

	default:
		return nil, fmt.Errorf("unknown analysis type: %s", req.Type)
	}

	// Build snapshot context if available
	if snapshotState != nil {
		snapshotCtx = prompt.BuildAnalysisContext(req, *snapshotState, profileReg, workloadReg)
	}

	fullPrompt := prompt.BuildAnalyzePrompt(req, analysisData, snapshotCtx)
	response, err := p.SingleChat(ctx, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI inference failed: %w", err)
	}

	recommendations := generateRecommendations(req.Type, anomalies, analysisData)

	baselineStatus := "normal"
	for _, a := range anomalies {
		if a.Severity == "high" || a.Severity == "critical" {
			baselineStatus = "elevated"
			break
		}
	}

	return &types.AnalyzeResponse{
		Summary:         response,
		Anomalies:       anomalies,
		BaselineStatus:  baselineStatus,
		Recommendations: recommendations,
		RelatedInsights: []types.RelatedInsight{},
	}, nil
}


func analyzeProcessAnomalies(profile *proc.ProcessProfile) []types.Anomaly {
	var out []types.Anomaly

	// Very simple heuristics: high IO or network counts are considered interesting.
	if profile.Dynamic.FileOpenCount > 1_000 {
		out = append(out, types.Anomaly{
			Type:        "unusual_file_activity",
			Description: "Process has opened an unusually high number of files.",
			Severity:    "high",
			Confidence:  0.7,
			Evidence: []string{
				fmt.Sprintf("FileOpenCount=%d", profile.Dynamic.FileOpenCount),
			},
		})
	}
	if profile.Dynamic.NetConnectCount > 1_000 {
		out = append(out, types.Anomaly{
			Type:        "unusual_network_activity",
			Description: "Process has made an unusually high number of network connections.",
			Severity:    "high",
			Confidence:  0.7,
			Evidence: []string{
				fmt.Sprintf("NetConnectCount=%d", profile.Dynamic.NetConnectCount),
			},
		})
	}
	return out
}


func analyzeWorkloadAnomalies(w *workload.Metadata) []types.Anomaly {
	var out []types.Anomaly
	totalEvents := w.ExecCount + w.FileCount + w.ConnectCount
	if totalEvents > 10_000 {
		out = append(out, types.Anomaly{
			Type:        "high_activity",
			Description: "Workload is generating a large volume of telemetry.",
			Severity:    "medium",
			Confidence:  0.6,
			Evidence: []string{
				fmt.Sprintf("TotalEvents=%d", totalEvents),
			},
		})
	}
	if w.AlertCount > 0 {
		out = append(out, types.Anomaly{
			Type:        "alerts_present",
			Description: "Workload has associated security alerts.",
			Severity:    "medium",
			Confidence:  0.8,
			Evidence: []string{
				fmt.Sprintf("AlertCount=%d", w.AlertCount),
			},
		})
	}
	return out
}


func analyzeRuleAnomalies(rule *rules.Rule, engine *rules.Engine) []types.Anomaly {
	var out []types.Anomaly
	testingBuffer := engine.GetTestingBuffer()
	if testingBuffer == nil {
		return out
	}
	stats := testingBuffer.GetStats(rule.Name)
	if !rule.IsTesting() {
		return out
	}
	if stats.Hits >= 100 {
		out = append(out, types.Anomaly{
			Type:        "testing_rule_hot",
			Description: "Testing rule is matching frequently and may be ready for promotion.",
			Severity:    "medium",
			Confidence:  0.8,
			Evidence: []string{
				fmt.Sprintf("Hits=%d", stats.Hits),
				fmt.Sprintf("ObservationMinutes=%d", stats.ObservationMinutes),
			},
		})
	}
	return out
}


func generateRecommendations(analysisType string, anomalies []types.Anomaly, data string) []types.Recommendation {
	if len(anomalies) == 0 {
		return nil
	}

	var recs []types.Recommendation

	switch analysisType {
	case "process", "workload":
		recs = append(recs, types.Recommendation{
			Type:        "investigation",
			Description: "Review the process/workload in Investigations to confirm whether behaviour is expected.",
			Priority:    "medium",
			Action: types.Action{
				Label:    "Investigate in UI",
				ActionID: "investigate",
				Params:   map[string]any{"context_type": analysisType},
			},
		})
	case "rule":
		recs = append(recs, types.Recommendation{
			Type:        "rule_creation",
			Description: "Consider promoting this testing rule or tightening its match criteria based on observed hits.",
			Priority:    "medium",
			Action: types.Action{
				Label:    "Review rule",
				ActionID: "review_rule",
				Params:   map[string]any{"rule_name": data},
			},
		})
	}

	return recs
}
