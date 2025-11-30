package rules

import (
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

func matchCgroupID(pattern string, cgroupID uint64) bool {
	return pattern == "" || strconv.FormatUint(cgroupID, 10) == pattern
}

func matchPID(pattern uint32, pid uint32) bool {
	return pattern == 0 || pid == pattern
}

// Returns: matched (any rule matched), rule (the matching rule), allowed (should the action be allowed)
func filterRulesByAction[T any](rules []*Rule, matchFn func(*Rule, T) bool, event T) (matched bool, rule *Rule, allowed bool) {
	var blockRule *Rule
	var alertRule *Rule

	for _, r := range rules {
		if !matchFn(r, event) {
			continue
		}
		switch r.Action {
		case ActionAllow:
			return true, r, true
		case ActionBlock:
			if blockRule == nil {
				blockRule = r
			}
		case ActionAlert:
			if alertRule == nil {
				alertRule = r
			}
		}
	}

	if blockRule != nil {
		return true, blockRule, false
	}
	if alertRule != nil {
		return true, alertRule, false
	}
	return false, nil, false
}
