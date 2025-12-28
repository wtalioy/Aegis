package snapshot

import (
	"strings"

	"aegis/pkg/apimodel"
	"aegis/pkg/proc"
)

// buildAncestorChainMaps creates maps of process keys and names to ancestor chains
func (s *Snapshot) buildAncestorChainMaps(execs []apimodel.ExecEvent, alerts []apimodel.Alert) (map[string]string, map[string]string) {
	processKeyToChain := make(map[string]string)
	processNameToChain := make(map[string]string)

	if s.processTree == nil {
		return processKeyToChain, processNameToChain
	}

	// Build map of process key (comm|parentComm) -> ancestor chain from exec events
	for _, exec := range execs {
		key := exec.Comm + "|" + exec.ParentComm
		if _, exists := processKeyToChain[key]; !exists {
			ancestors := s.processTree.GetAncestors(exec.PID)
			if len(ancestors) > 0 {
				chain := formatAncestorChain(ancestors)
				processKeyToChain[key] = chain
				// Also map by process name for alerts
				if exec.Comm != "" {
					processNameToChain[exec.Comm] = chain
				}
			}
		}
	}

	// Also build chains for processes in alerts (by PID if available)
	for _, alert := range alerts {
		if alert.PID != 0 {
			ancestors := s.processTree.GetAncestors(alert.PID)
			if len(ancestors) > 0 {
				chain := formatAncestorChain(ancestors)
				if alert.ProcessName != "" {
					processNameToChain[alert.ProcessName] = chain
				}
			}
		}
	}

	return processKeyToChain, processNameToChain
}

// formatAncestorChain formats a process ancestor chain as a readable string.
func formatAncestorChain(ancestors []*proc.ProcessInfo) string {
	if len(ancestors) == 0 {
		return ""
	}

	// Limit depth
	maxDepth := MaxAncestorDepth
	if len(ancestors) > maxDepth {
		ancestors = ancestors[:maxDepth]
	}

	parts := make([]string, 0, len(ancestors))
	for i := len(ancestors) - 1; i >= 0; i-- {
		if ancestors[i].Comm != "" {
			parts = append(parts, ancestors[i].Comm)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " → ")
}

