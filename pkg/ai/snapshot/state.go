package snapshot

import (
	"fmt"
	"sort"
	"time"

	"aegis/pkg/apimodel"
)

// buildBaseState creates the base system state with metrics
func (s *Snapshot) buildBaseState(execRate, fileRate, netRate int64) SystemState {
	totalRate := execRate + fileRate + netRate
	processCount := 0
	if s.processTree != nil {
		processCount = s.processTree.Size()
	}

	return SystemState{
		Timestamp:     time.Now(),
		LoadLevel:     calculateLoadLevel(totalRate),
		ExecRate:      execRate,
		FileRate:      fileRate,
		NetworkRate:   netRate,
		ProcessCount:  processCount,
		WorkloadCount: s.stats.WorkloadCount(),
		AlertCount:    int(s.stats.TotalAlertCount()),
	}
}

// calculateLoadLevel determines the load level based on total event rate
func calculateLoadLevel(totalRate int64) string {
	switch {
	case totalRate > 1000:
		return "critical"
	case totalRate > 500:
		return "high"
	case totalRate < 50:
		return "low"
	default:
		return "normal"
	}
}

// buildTopWorkloads extracts and sorts the top 5 workloads by total event count
func (s *Snapshot) buildTopWorkloads() []WorkloadSummary {
	if s.workloadReg == nil {
		return nil
	}

	workloads := s.workloadReg.List()
	sort.Slice(workloads, func(i, j int) bool {
		totalI := workloads[i].ExecCount + workloads[i].FileCount + workloads[i].ConnectCount
		totalJ := workloads[j].ExecCount + workloads[j].FileCount + workloads[j].ConnectCount
		return totalI > totalJ
	})

	topCount := 5
	if len(workloads) < topCount {
		topCount = len(workloads)
	}

	result := make([]WorkloadSummary, 0, topCount)
	for i := 0; i < topCount; i++ {
		w := workloads[i]
		result = append(result, WorkloadSummary{
			ID:          fmt.Sprintf("%d", w.ID),
			CgroupPath:  w.CgroupPath,
			TotalEvents: w.ExecCount + w.FileCount + w.ConnectCount,
			AlertCount:  w.AlertCount,
		})
	}
	return result
}

// deduplicateAlerts groups alerts by rule name and process name, aggregating counts
func (s *Snapshot) deduplicateAlerts(alerts []apimodel.Alert) []AlertSummary {
	groups := make(map[string]*AlertSummary)

	for _, alert := range alerts {
		key := alert.RuleName + "|" + alert.ProcessName
		if existing, ok := groups[key]; ok {
			existing.Count++
			if alert.Blocked {
				existing.WasBlocked = true
			}
		} else {
			groups[key] = &AlertSummary{
				RuleName:    alert.RuleName,
				Severity:    alert.Severity,
				ProcessName: alert.ProcessName,
				Count:       1,
				WasBlocked:  alert.Blocked,
			}
		}
	}

	return finalizeGroup(groups, MaxAlertSummaries, func(a, b AlertSummary) bool {
		if a.Severity != b.Severity {
			return severityOrder(a.Severity) > severityOrder(b.Severity)
		}
		return a.Count > b.Count
	})
}

