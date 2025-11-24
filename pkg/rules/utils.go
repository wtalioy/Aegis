package rules

import (
	"net"
	"strings"
)

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

func matchIP(eventIP, ruleIP string) bool {
	_, ipNet, err := net.ParseCIDR(ruleIP)
	if err == nil {
		ip := net.ParseIP(eventIP)
		return ip != nil && ipNet.Contains(ip)
	}
	return eventIP == ruleIP
}
