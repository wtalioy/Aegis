package rules

import (
	"eulerguard/pkg/events"
	"eulerguard/pkg/utils"
)

type connectMatcher struct {
	rules []*Rule
}

func newConnectMatcher(rules []Rule) *connectMatcher {
	matcher := &connectMatcher{
		rules: make([]*Rule, 0),
	}
	for i := range rules {
		rule := &rules[i]
		if rule.Match.DestPort != 0 || rule.Match.DestIP != "" {
			matcher.rules = append(matcher.rules, rule)
		}
	}
	return matcher
}

func (m *connectMatcher) Match(event *events.ConnectEvent) (bool, *Rule) {
	for _, rule := range m.rules {
		if m.matchRule(rule, event) {
			return true, rule
		}
	}
	return false, nil
}

func (m *connectMatcher) matchRule(rule *Rule, event *events.ConnectEvent) bool {
	match := rule.Match

	if match.DestPort == 0 && match.DestIP == "" {
		return false
	}
	if match.DestPort != 0 && event.Port != match.DestPort {
		return false
	}
	if match.DestIP != "" {
		eventIP := utils.ExtractIP(event)
		if eventIP == "" || !matchIP(eventIP, match.DestIP) {
			return false
		}
	}
	if match.InContainer && event.CgroupID == 1 {
		return false
	}
	if match.PID != 0 && event.PID != match.PID {
		return false
	}
	return true
}
