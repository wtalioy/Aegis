package output

import (
	"eulerguard/pkg/events"
	"eulerguard/pkg/proctree"
	"eulerguard/pkg/utils"
	"fmt"
	"strings"
)

const ansiReset = "\033[0m"

func formatAlertText(ruleName, severity, description string, pid uint32, process string, ppid uint32, parent string, cgroupID uint64) string {
	return fmt.Sprintf("[Alert!] Rule '%s' triggered [Severity: %s]\n"+
		"  Description: %s\n"+
		"  Process: PID=%d(%s) ← PPID=%d(%s) | Cgroup=%d\n",
		ruleName, severity, description,
		pid, process, ppid, parent, cgroupID)
}

func formatFileAlertText(ruleName, severity, description, filename string,
	pid uint32, cgroupID uint64, flags uint32, chain []*proctree.ProcessInfo) string {

	var builder strings.Builder
	fmt.Fprintf(&builder, "[ALERT!] Rule '%s' triggered [Severity: %s]\n", ruleName, severity)
	fmt.Fprintf(&builder, "  Description: %s\n", description)
	fmt.Fprintf(&builder, "  File: %s\n", filename)
	fmt.Fprintf(&builder, "  PID: %d | Cgroup: %d | Flags: 0x%x\n", pid, cgroupID, flags)

	if len(chain) > 0 {
		fmt.Fprintf(&builder, "  Attack Chain: %s\n", formatChain(chain))
	}

	return builder.String()
}

func formatConnectAlertText(ruleName, severity, description, destAddr string,
	pid uint32, cgroupID uint64, chain []*proctree.ProcessInfo) string {

	var builder strings.Builder
	fmt.Fprintf(&builder, "[ALERT!] Rule '%s' triggered [Severity: %s]\n", ruleName, severity)
	fmt.Fprintf(&builder, "  Description: %s\n", description)
	fmt.Fprintf(&builder, "  Network Connection: %s\n", destAddr)
	fmt.Fprintf(&builder, "  PID: %d | Cgroup: %d\n", pid, cgroupID)

	if len(chain) > 0 {
		fmt.Fprintf(&builder, "  Attack Chain: %s\n", formatChain(chain))
	}

	return builder.String()
}

func formatAddress(ev *events.ConnectEvent) string {
	ip := utils.ExtractIP(ev)
	switch ev.Family {
	case 10:
		return fmt.Sprintf("[%s]:%d", ip, ev.Port)
	case 2:
		return fmt.Sprintf("%s:%d", ip, ev.Port)
	default:
		return fmt.Sprintf("unknown_family_%d:port_%d", ev.Family, ev.Port)
	}
}

func formatChain(chain []*proctree.ProcessInfo) string {
	parts := make([]string, len(chain))
	for i, info := range reverseChain(chain) {
		parts[i] = fmt.Sprintf("%s(%d)", info.Comm, info.PID)
	}
	return strings.Join(parts, " -> ")
}

func formatChainJSON(chain []*proctree.ProcessInfo) []map[string]any {
	result := make([]map[string]any, len(chain))
	for i, info := range reverseChain(chain) {
		result[i] = map[string]any{
			"pid":       info.PID,
			"ppid":      info.PPID,
			"comm":      info.Comm,
			"cgroup_id": info.CgroupID,
		}
	}
	return result
}

func reverseChain(chain []*proctree.ProcessInfo) []*proctree.ProcessInfo {
	reversed := make([]*proctree.ProcessInfo, len(chain))
	for i, info := range chain {
		reversed[len(chain)-1-i] = info
	}
	return reversed
}

func getSeverityColor(severity string) string {
	switch severity {
	case "high", "critical":
		return "\033[1;31m" // Bold Red
	case "warning", "medium":
		return "\033[1;33m" // Bold Yellow
	case "info", "low":
		return "\033[1;36m" // Bold Cyan
	default:
		return "\033[1;37m" // Bold White
	}
}
