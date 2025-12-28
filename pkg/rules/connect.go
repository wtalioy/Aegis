package rules

import (
	"aegis/pkg/events"
	"aegis/pkg/utils"
)

type connectMatcher struct {
	rules []*Rule
}

func newConnectMatcher(rules []Rule) *connectMatcher {
	matcher := &connectMatcher{rules: make([]*Rule, 0)}
	for i := range rules {
		if rules[i].Match.DestPort != 0 || rules[i].Match.DestIP != "" {
			matcher.rules = append(matcher.rules, &rules[i])
		}
	}
	return matcher
}

func (m *connectMatcher) Match(event *events.ConnectEvent) (matched bool, rule *Rule, allowed bool) {
	return filterRulesByAction(m.rules, m.matchRule, event)
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
		if eventIP := utils.ExtractIP(event); eventIP == "" || !match.MatchIP(eventIP) {
			return false
		}
	}
	return matchCgroupID(match.CgroupID, event.Hdr.CgroupID) && matchPID(match.PID, event.Hdr.PID)
}
