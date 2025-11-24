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
	matcher.indexRules(rules)
	return matcher
}

func (m *execMatcher) indexRules(rules []Rule) {
	for i := range rules {
		rule := &rules[i]
		m.setDefaultMatchTypes(rule)
		if !m.hasExecCriteria(rule) {
			continue
		}
		m.indexExecRule(rule)
	}
}

func (m *execMatcher) setDefaultMatchTypes(rule *Rule) {
	if rule.Match.ProcessName != "" && rule.Match.ProcessNameType == "" {
		rule.Match.ProcessNameType = MatchTypeContains
	}
	if rule.Match.ParentName != "" && rule.Match.ParentNameType == "" {
		rule.Match.ParentNameType = MatchTypeContains
	}
}

func (m *execMatcher) hasExecCriteria(rule *Rule) bool {
	match := rule.Match
	return match.ProcessName != "" || match.ParentName != "" ||
		match.PID != 0 || match.PPID != 0
}

func (m *execMatcher) indexExecRule(rule *Rule) {
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

	needsPartialMatch := !indexed ||
		rule.Match.ProcessNameType == MatchTypeContains ||
		rule.Match.ParentNameType == MatchTypeContains

	if needsPartialMatch {
		m.partialMatchRules = append(m.partialMatchRules, rule)
	}
}

func (m *execMatcher) Match(event events.ProcessedEvent) []Alert {
	var alerts []Alert
	checked := make(map[*Rule]bool)

	if rules, ok := m.exactProcessNameRules[event.Process]; ok {
		m.applyRules(rules, event, checked, &alerts)
	}
	if rules, ok := m.exactParentNameRules[event.Parent]; ok {
		m.applyRules(rules, event, checked, &alerts)
	}

	m.applyRules(m.partialMatchRules, event, checked, &alerts)
	return alerts
}

func (m *execMatcher) applyRules(rules []*Rule, event events.ProcessedEvent, checked map[*Rule]bool, alerts *[]Alert) {
	for _, rule := range rules {
		if checked[rule] {
			continue
		}
		if m.matchRule(rule, event) {
			*alerts = append(*alerts, Alert{
				Rule:    *rule,
				Event:   event,
				Message: rule.Description,
			})
			checked[rule] = true
		}
	}
}

func (m *execMatcher) matchRule(rule *Rule, event events.ProcessedEvent) bool {
	match := rule.Match
	return (match.ProcessName == "" || matchString(event.Process, match.ProcessName, match.ProcessNameType)) &&
		(match.ParentName == "" || matchString(event.Parent, match.ParentName, match.ParentNameType)) &&
		(match.PID == 0 || event.Event.PID == match.PID) &&
		(match.PPID == 0 || event.Event.PPID == match.PPID) &&
		(!match.InContainer || event.Event.CgroupID != 1)
}
