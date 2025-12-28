package snapshot

import (
	"sort"
	"strings"
)

// compareByBlockedThenCount compares two items prioritizing blocked status, then count
func compareByBlockedThenCount(blockedA bool, countA int, blockedB bool, countB int) bool {
	if blockedA != blockedB {
		return blockedA
	}
	return countA > countB
}

// simplifyFilePath simplifies file paths for display
func simplifyFilePath(path string) string {
	// Keep full paths for important directories
	if strings.HasPrefix(path, "/etc/") || strings.HasPrefix(path, "/root/") || strings.HasPrefix(path, "/home/") {
		return path
	}

	parts := strings.Split(path, "/")
	if len(parts) <= 3 {
		return path
	}

	// Simplify /proc paths
	if strings.HasPrefix(path, "/proc/") {
		return "/proc/[pid]/" + strings.Join(parts[3:], "/")
	}

	// Simplify /tmp and /var paths
	if strings.HasPrefix(path, "/tmp/") || strings.HasPrefix(path, "/var/") {
		return "/" + parts[1] + "/" + parts[2] + "/..."
	}

	return path
}


func finalizeGroup[T any](groups map[string]*T, limit int, less func(a, b T) bool) []T {
	result := make([]T, 0, len(groups))
	for _, v := range groups {
		result = append(result, *v)
	}

	sort.Slice(result, func(i, j int) bool {
		return less(result[i], result[j])
	})

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result
}

// severityOrder returns a numeric order for severity levels
func severityOrder(severity string) int {
	switch severity {
	case "critical":
		return 4
	case "high":
		return 3
	case "warning":
		return 2
	default:
		return 1
	}
}

