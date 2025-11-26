package rules

import "eulerguard/pkg/events"

type execMatcher struct {
	exactProcessNameRules map[string][]*Rule
	exactParentNameRules  map[string][]*Rule
	partialMatchRules     []*Rule
}

func newExecMatcher(rules []Rule) *execMatcher {
	matcher := &execMatcher{
		exactProcessNameRules: make(map[string][]*Rule),
		exactParentNameRules:  make(map[string][]*Rule),
		partialMatchRules:     make([]*Rule, 0),
	}
	for i := range rules {
		rule := &rules[i]
		setDefaultMatchTypes(rule)
		if hasExecCriteria(rule) {
			matcher.indexRule(rule)
		}
	}
	return matcher
}

func setDefaultMatchTypes(rule *Rule) {
	if rule.Match.ProcessName != "" && rule.Match.ProcessNameType == "" {
		rule.Match.ProcessNameType = MatchTypeContains
	}
	if rule.Match.ParentName != "" && rule.Match.ParentNameType == "" {
		rule.Match.ParentNameType = MatchTypeContains
	}
}

func hasExecCriteria(rule *Rule) bool {
	m := rule.Match
	return m.ProcessName != "" || m.ParentName != "" || m.PID != 0 || m.PPID != 0
}

func (m *execMatcher) indexRule(rule *Rule) {
	indexed := false
	if rule.Match.ProcessName != "" && rule.Match.ProcessNameType == MatchTypeExact {
		m.exactProcessNameRules[rule.Match.ProcessName] = append(
			m.exactProcessNameRules[rule.Match.ProcessName], rule)
		indexed = true
	}
	if rule.Match.ParentName != "" && rule.Match.ParentNameType == MatchTypeExact {
		m.exactParentNameRules[rule.Match.ParentName] = append(
			m.exactParentNameRules[rule.Match.ParentName], rule)
		indexed = true
	}
	if !indexed || rule.Match.ProcessNameType == MatchTypeContains || rule.Match.ParentNameType == MatchTypeContains {
		m.partialMatchRules = append(m.partialMatchRules, rule)
	}
}

func (m *execMatcher) Match(event events.ProcessedEvent) (matched bool, rule *Rule, allowed bool) {
	return filterRulesByAction(m.getCandidateRules(event), m.matchRuleWrapper, event)
}

func (m *execMatcher) matchRuleWrapper(rule *Rule, event events.ProcessedEvent) bool {
	return m.matchRule(rule, event)
}

func (m *execMatcher) CollectAlerts(event events.ProcessedEvent) []Alert {
	candidates := m.getCandidateRules(event)
	for _, rule := range candidates {
		if rule.Action == ActionAllow && m.matchRule(rule, event) {
			return nil
		}
	}
	seen := make(map[*Rule]bool)
	var alerts []Alert
	for _, rule := range candidates {
		if seen[rule] || rule.Action == ActionAllow {
			continue
		}
		seen[rule] = true
		if m.matchRule(rule, event) {
			alerts = append(alerts, Alert{Rule: *rule, Event: event, Message: rule.Description})
		}
	}
	return alerts
}

func (m *execMatcher) getCandidateRules(event events.ProcessedEvent) []*Rule {
	var candidates []*Rule
	if rules, ok := m.exactProcessNameRules[event.Process]; ok {
		candidates = append(candidates, rules...)
	}
	if rules, ok := m.exactParentNameRules[event.Parent]; ok {
		candidates = append(candidates, rules...)
	}
	candidates = append(candidates, m.partialMatchRules...)
	return candidates
}

func (m *execMatcher) matchRule(rule *Rule, event events.ProcessedEvent) bool {
	match := rule.Match
	return (match.ProcessName == "" || matchString(event.Process, match.ProcessName, match.ProcessNameType)) &&
		(match.ParentName == "" || matchString(event.Parent, match.ParentName, match.ParentNameType)) &&
		matchPID(match.PID, event.Event.PID) &&
		(match.PPID == 0 || event.Event.PPID == match.PPID) &&
		matchCgroupID(match.CgroupID, event.Event.CgroupID)
}
