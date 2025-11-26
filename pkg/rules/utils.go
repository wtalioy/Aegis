package rules

import (
	"net"
	"strconv"
	"strings"
)

// match a value against a pattern using the specified match type.
func matchString(value, pattern string, matchType MatchType) bool {
	switch matchType {
	case MatchTypeExact:
		return value == pattern
	case MatchTypePrefix:
		return strings.HasPrefix(value, pattern)
	case MatchTypeContains:
		return strings.Contains(value, pattern)
	default:
		return strings.Contains(value, pattern)
	}
}

// match an IP address against a pattern (supports CIDR notation).
func matchIP(eventIP, ruleIP string) bool {
	_, ipNet, err := net.ParseCIDR(ruleIP)
	if err == nil {
		ip := net.ParseIP(eventIP)
		return ip != nil && ipNet.Contains(ip)
	}
	return eventIP == ruleIP
}

func matchCgroupID(pattern string, cgroupID uint64) bool {
	return pattern == "" || strconv.FormatUint(cgroupID, 10) == pattern
}

func matchPID(pattern uint32, pid uint32) bool {
	return pattern == 0 || pid == pattern
}

func filterRulesByAction[T any](rules []*Rule, matchFn func(*Rule, T) bool, event T) (matched bool, rule *Rule, allowed bool) {
	for _, r := range rules {
		if r.Action == ActionAllow && matchFn(r, event) {
			return true, r, true
		}
	}
	for _, r := range rules {
		if r.Action != ActionAllow && matchFn(r, event) {
			return true, r, false
		}
	}
	return false, nil, false
}
