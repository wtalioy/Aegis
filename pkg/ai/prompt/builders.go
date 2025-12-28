package prompt

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"aegis/pkg/ai/snapshot"
	"aegis/pkg/ai/types"
	"aegis/pkg/events"
	"aegis/pkg/proc"
	"aegis/pkg/rules"
	"aegis/pkg/storage"
	"aegis/pkg/utils"
	"aegis/pkg/workload"
)

func BuildAnalyzePrompt(req *types.AnalyzeRequest, analysisData string, snapshotContext string) string {
	var user strings.Builder
	user.WriteString("Analysis request:\n")
	user.WriteString(fmt.Sprintf("- Type: %s\n", req.Type))
	user.WriteString(fmt.Sprintf("- ID: %s\n\n", req.ID))
	
	if snapshotContext != "" {
		user.WriteString("System Context:\n")
		user.WriteString(snapshotContext)
		user.WriteString("\n\n")
	}
	
	user.WriteString("Target Entity:\n")
	user.WriteString(analysisData)
	user.WriteString("\n\nAnalyze:")
	
	return AnalyzeSystemPrompt + "\n\n" + user.String()
}

func BuildAnalysisContext(req *types.AnalyzeRequest, state snapshot.SystemState, profileReg *proc.ProfileRegistry, workloadReg *workload.Registry) string {
	var b strings.Builder
	
	// Always include system load level
	b.WriteString(fmt.Sprintf("System Load: %s (Exec: %d/s, File: %d/s, Network: %d/s)\n", 
		state.LoadLevel, state.ExecRate, state.FileRate, state.NetworkRate))
	
	switch req.Type {
	case types.AnalyzeTypeProcess:
		pid, err := strconv.ParseUint(req.ID, 10, 32)
		if err == nil {
			// Include related alerts for this process
			if len(state.RecentAlerts) > 0 {
				alertCount := 0
				for _, alert := range state.RecentAlerts {
					if alert.ProcessName != "" && alertCount < 3 {
						// Try to match by PID if we have profile
						if profile, ok := profileReg.GetProfile(uint32(pid)); ok {
							comm := profile.Static.CommandLine
							if comm != "" && (strings.Contains(comm, alert.ProcessName) || strings.Contains(alert.ProcessName, comm)) {
								b.WriteString(fmt.Sprintf("Related Alert: %s (%s) - %d occurrence(s)\n", 
									alert.RuleName, alert.Severity, alert.Count))
								alertCount++
							}
						}
					}
				}
			}
			// Include recent process activity
			if len(state.RecentProcesses) > 0 {
				b.WriteString("\nRecent Process Activity:\n")
				for i, proc := range state.RecentProcesses {
					if i >= 3 {
						break
					}
					b.WriteString(fmt.Sprintf("- %s → %s (%d times)\n", 
						proc.ParentComm, proc.Comm, proc.Count))
				}
			}
		}
		
	case types.AnalyzeTypeWorkload:
		cgroupID, err := strconv.ParseUint(req.ID, 10, 64)
		if err == nil {
			w := workloadReg.Get(cgroupID)
			if w != nil {
				// Include workload-level alerts
				if len(state.RecentAlerts) > 0 {
					b.WriteString("\nWorkload Alerts:\n")
					for i, alert := range state.RecentAlerts {
						if i >= 5 {
							break
						}
						b.WriteString(fmt.Sprintf("- %s: %s (%s) - %d occurrence(s)\n", 
							alert.ProcessName, alert.RuleName, alert.Severity, alert.Count))
					}
				}
			}
		}
		
	case types.AnalyzeTypeRule:
		// Include recent alerts matching this rule
		if len(state.RecentAlerts) > 0 {
			b.WriteString("\nRecent Rule Matches:\n")
			for _, alert := range state.RecentAlerts {
				if alert.RuleName == req.ID {
					b.WriteString(fmt.Sprintf("- Process: %s, Severity: %s, Count: %d\n", 
						alert.ProcessName, alert.Severity, alert.Count))
				}
			}
		}
	}
	
	return b.String()
}

// Helpers reused by analysis/analyze.go.
func FormatProcessProfile(profile *proc.ProcessProfile) string {
	return fmt.Sprintf("PID: %d, ExecCount: %d, FileCount: %d, NetCount: %d",
		profile.PID, profile.Dynamic.ExecCount, profile.Dynamic.FileOpenCount, profile.Dynamic.NetConnectCount)
}

func FormatWorkloadMetadata(w *workload.Metadata) string {
	return fmt.Sprintf("CgroupID: %d, CgroupPath: %s, ExecCount: %d, FileCount: %d, ConnectCount: %d, AlertCount: %d",
		w.ID, w.CgroupPath, w.ExecCount, w.FileCount, w.ConnectCount, w.AlertCount)
}

func FormatRuleForAnalysis(rule *rules.Rule, engine *rules.Engine) string {
	testingBuffer := engine.GetTestingBuffer()
	stats := testingBuffer.GetStats(rule.Name)
	return fmt.Sprintf("Rule: %s, Mode: %s, Action: %s, Hits: %d",
		rule.Name, rule.State, rule.Action, stats.Hits)
}

func BuildExplainPrompt(eventDesc, question string, relatedContext string) string {
	if strings.TrimSpace(question) == "" {
		question = "Explain this event"
	}
	var user strings.Builder
	user.WriteString("Event details:\n")
	user.WriteString(eventDesc)
	user.WriteString("\n\n")
	
	if relatedContext != "" {
		user.WriteString("Related Context:\n")
		user.WriteString(relatedContext)
		user.WriteString("\n\n")
	}
	
	user.WriteString(fmt.Sprintf("User question: \"%s\"\n\nExplain:", question))
	return ExplainSystemPrompt + "\n\n" + user.String()
}

func BuildExplainContext(event *storage.Event, pid uint32, store storage.EventStore, profileReg *proc.ProfileRegistry, workloadReg *workload.Registry, processTree *proc.ProcessTree, state *snapshot.SystemState) string {
	var b strings.Builder
	
	// Process profile context
	if profileReg != nil && pid != 0 {
		if profile, ok := profileReg.GetProfile(pid); ok {
			b.WriteString("Process Profile:\n")
			b.WriteString(fmt.Sprintf("- PID: %d\n", profile.PID))
			if profile.Static.CommandLine != "" {
				b.WriteString(fmt.Sprintf("- Command: %s\n", profile.Static.CommandLine))
			}
			b.WriteString(fmt.Sprintf("- Activity: Exec=%d, File=%d, Net=%d\n", 
				profile.Dynamic.ExecCount, profile.Dynamic.FileOpenCount, profile.Dynamic.NetConnectCount))
			
			// Process ancestry
			if processTree != nil {
				ancestors := processTree.GetAncestors(pid)
				if len(ancestors) > 0 {
					b.WriteString("- Ancestry: ")
					parts := make([]string, 0, len(ancestors))
					maxDepth := 5
					if len(ancestors) > maxDepth {
						ancestors = ancestors[:maxDepth]
					}
					for i := len(ancestors) - 1; i >= 0; i-- {
						if ancestors[i].Comm != "" {
							parts = append(parts, ancestors[i].Comm)
						}
					}
					if len(parts) > 0 {
						b.WriteString(strings.Join(parts, " → "))
					}
					b.WriteString("\n")
				}
			}
			b.WriteString("\n")
		}
	}
	
	// Workload context
	if workloadReg != nil && pid != 0 {
		if processTree != nil {
			if procInfo, ok := processTree.GetProcess(pid); ok && procInfo != nil {
				cgroupID := procInfo.CgroupID
				if w := workloadReg.Get(cgroupID); w != nil {
					b.WriteString("Workload Context:\n")
					b.WriteString(fmt.Sprintf("- Cgroup: %s\n", w.CgroupPath))
					b.WriteString(fmt.Sprintf("- Events: Exec=%d, File=%d, Connect=%d\n", 
						w.ExecCount, w.FileCount, w.ConnectCount))
					if w.AlertCount > 0 {
						b.WriteString(fmt.Sprintf("- Alerts: %d\n", w.AlertCount))
					}
					b.WriteString("\n")
				}
			}
		}
	}
	
	// Related events from storage (same process, recent time window)
	if store != nil && pid != 0 {
		now := time.Now()
		windowStart := now.Add(-5 * time.Minute)
		eventList, err := store.Query(windowStart, now)
		if err == nil && len(eventList) > 0 {
			relatedCount := 0
			b.WriteString("Related Recent Events:\n")
			for _, ev := range eventList {
				if ev == nil || ev.Timestamp.Before(windowStart) {
					continue
				}
				var evPID uint32
				var eventTypeStr string
				switch e := ev.Data.(type) {
				case *events.ExecEvent:
					evPID = e.Hdr.PID
					eventTypeStr = "exec"
				case *events.FileOpenEvent:
					evPID = e.Hdr.PID
					eventTypeStr = "file"
				case *events.ConnectEvent:
					evPID = e.Hdr.PID
					eventTypeStr = "connect"
				default:
					// Try to infer from storage.Event.Type
					switch ev.Type {
					case events.EventTypeExec:
						eventTypeStr = "exec"
					case events.EventTypeFileOpen:
						eventTypeStr = "file"
					case events.EventTypeConnect:
						eventTypeStr = "connect"
					default:
						eventTypeStr = "unknown"
					}
					// Try to extract PID from map[string]any if available
					if m, ok := ev.Data.(map[string]any); ok {
						if hdr, ok := m["header"].(map[string]any); ok {
							if v, ok := hdr["pid"].(float64); ok {
								evPID = uint32(v)
							}
						}
						if v, ok := m["pid"].(float64); ok {
							evPID = uint32(v)
						}
					}
				}
				if evPID == pid && evPID != 0 && relatedCount < 5 {
					b.WriteString(fmt.Sprintf("- %s: %s\n", ev.Timestamp.Format("15:04:05"), eventTypeStr))
					relatedCount++
				}
			}
			if relatedCount == 0 {
				b.WriteString("- None in last 5 minutes\n")
			}
			b.WriteString("\n")
		}
	}
	
	// System-level alerts related to this process
	if state != nil && len(state.RecentAlerts) > 0 {
		alertCount := 0
		for _, alert := range state.RecentAlerts {
			if alert.ProcessName != "" && alertCount < 3 {
				b.WriteString(fmt.Sprintf("System Alert: %s (%s) - %d occurrence(s)\n", 
					alert.RuleName, alert.Severity, alert.Count))
				alertCount++
			}
		}
	}
	
	return b.String()
}

func FormatEventForExplanation(event *storage.Event, profile *proc.ProcessProfile) string {
	var b strings.Builder
	b.WriteString("Event\n")
	b.WriteString(fmt.Sprintf("- Time: %s\n", event.Timestamp.Format(time.RFC3339)))

	switch ev := event.Data.(type) {
	case *events.ExecEvent:
		b.WriteString("- Type: exec\n")
		b.WriteString(fmt.Sprintf("- PID: %d\n", ev.Hdr.PID))
		b.WriteString(fmt.Sprintf("- PPID: %d\n", ev.PPID))
		b.WriteString(fmt.Sprintf("- Comm: %s\n", strings.TrimRight(string(ev.Hdr.Comm[:]), "\x00")))
		b.WriteString(fmt.Sprintf("- ParentComm: %s\n", strings.TrimRight(string(ev.PComm[:]), "\x00")))
	case *events.FileOpenEvent:
		b.WriteString("- Type: file\n")
		b.WriteString(fmt.Sprintf("- PID: %d\n", ev.Hdr.PID))
		b.WriteString(fmt.Sprintf("- Comm: %s\n", strings.TrimRight(string(ev.Hdr.Comm[:]), "\x00")))
		b.WriteString(fmt.Sprintf("- Filename: %s\n", utils.ExtractCString(ev.Filename[:])))
		b.WriteString(fmt.Sprintf("- Flags: %d, Dev: %d, Ino: %d\n", ev.Flags, ev.Dev, ev.Ino))
	case *events.ConnectEvent:
		b.WriteString("- Type: connect\n")
		b.WriteString(fmt.Sprintf("- PID: %d\n", ev.Hdr.PID))
		b.WriteString(fmt.Sprintf("- Comm: %s\n", strings.TrimRight(string(ev.Hdr.Comm[:]), "\x00")))
		ip := utils.ExtractIP(ev)
		b.WriteString(fmt.Sprintf("- Remote: %s:%d (family=%d)\n", ip, ev.Port, ev.Family))
	case map[string]any:
		typeStr, _ := ev["type"].(string)
		if typeStr == "" {
			switch event.Type {
			case events.EventTypeExec:
				typeStr = "exec"
			case events.EventTypeFileOpen:
				typeStr = "file"
			case events.EventTypeConnect:
				typeStr = "connect"
			}
		}
		b.WriteString(fmt.Sprintf("- Type: %s\n", typeStr))
		if hdr, ok := ev["header"].(map[string]any); ok {
			if v, ok := hdr["pid"].(float64); ok {
				b.WriteString(fmt.Sprintf("- PID: %d\n", uint32(v)))
			}
			if comm, ok := hdr["comm"].(string); ok {
				b.WriteString(fmt.Sprintf("- Comm: %s\n", comm))
			}
		}
		if v, ok := ev["pid"].(float64); ok {
			b.WriteString(fmt.Sprintf("- PID: %d\n", uint32(v)))
		}
		if comm, ok := ev["comm"].(string); ok {
			b.WriteString(fmt.Sprintf("- Comm: %s\n", comm))
		}
		if parentComm, ok := ev["parentComm"].(string); ok {
			b.WriteString(fmt.Sprintf("- ParentComm: %s\n", parentComm))
		}
		if filename, ok := ev["filename"].(string); ok && filename != "" {
			b.WriteString(fmt.Sprintf("- Filename: %s\n", filename))
		}
		if addr, ok := ev["addr"].(string); ok && addr != "" {
			port := 0
			if p, ok := ev["port"].(float64); ok {
				port = int(p)
			}
			procName := ""
			if pn, ok := ev["processName"].(string); ok {
				procName = pn
			}
			if procName != "" {
				b.WriteString(fmt.Sprintf("- Process: %s\n", procName))
			}
			b.WriteString(fmt.Sprintf("- Remote: %s:%d\n", addr, port))
		}
	default:
		b.WriteString(fmt.Sprintf("- Type: %d\n", int(event.Type)))
	}

	if profile != nil {
		b.WriteString("\nProcess Profile\n")
		if !profile.Static.StartTime.IsZero() {
			b.WriteString(fmt.Sprintf("- StartTime: %s\n", profile.Static.StartTime.Format(time.RFC3339)))
		}
		if profile.Static.CommandLine != "" {
			b.WriteString(fmt.Sprintf("- CommandLine: %s\n", profile.Static.CommandLine))
		}
		b.WriteString(fmt.Sprintf("- ExecCount: %d, FileOpenCount: %d, NetConnectCount: %d\n", profile.Dynamic.ExecCount, profile.Dynamic.FileOpenCount, profile.Dynamic.NetConnectCount))
		if !profile.Dynamic.LastExec.IsZero() {
			b.WriteString(fmt.Sprintf("- LastExec: %s\n", profile.Dynamic.LastExec.Format(time.RFC3339)))
		}
		if !profile.Dynamic.LastFileOpen.IsZero() {
			b.WriteString(fmt.Sprintf("- LastFileOpen: %s\n", profile.Dynamic.LastFileOpen.Format(time.RFC3339)))
		}
		if !profile.Dynamic.LastConnect.IsZero() {
			b.WriteString(fmt.Sprintf("- LastConnect: %s\n", profile.Dynamic.LastConnect.Format(time.RFC3339)))
		}
	}

	return b.String()
}

func BuildRuleGenPrompt(req *types.RuleGenRequest, examplesYAML string) string {
	user := fmt.Sprintf(
		"User request: \"%s\"\n\nExisting rules (examples):\n%s\n\nGenerate rule:",
		req.Description,
		strings.TrimSpace(examplesYAML),
	)
	return RuleGenSystemPrompt + "\n\n" + user
}


