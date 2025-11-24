package rules

import (
	"eulerguard/pkg/events"
	"eulerguard/pkg/utils"
	"net"
	"strings"
)

type Engine struct {
	rules []Rule

	// Indexed lookups for exact matches
	exactProcessNameRules map[string][]*Rule
	exactParentNameRules  map[string][]*Rule
	exactFilenameRules    map[string][]*Rule

	// Non-exact matches
	partialMatchRules []*Rule
	filepathRules     []*Rule
}

func NewEngine(rules []Rule) *Engine {
	e := &Engine{
		rules:                 rules,
		exactProcessNameRules: make(map[string][]*Rule),
		exactParentNameRules:  make(map[string][]*Rule),
		exactFilenameRules:    make(map[string][]*Rule),
		partialMatchRules:     make([]*Rule, 0),
		filepathRules:         make([]*Rule, 0),
	}
	e.buildIndexes()
	return e
}

func (e *Engine) buildIndexes() {
	for i := range e.rules {
		rule := &e.rules[i]
		e.setDefaultMatchTypes(rule)

		// Index exec event rules
		if e.hasExecCriteria(rule) {
			e.indexExecRule(rule)
		}

		// Index file access rules
		if rule.Match.Filename != "" {
			e.exactFilenameRules[rule.Match.Filename] = append(
				e.exactFilenameRules[rule.Match.Filename], rule)
		}
		if rule.Match.FilePath != "" {
			e.filepathRules = append(e.filepathRules, rule)
		}
	}
}

func (e *Engine) setDefaultMatchTypes(rule *Rule) {
	if rule.Match.ProcessName != "" && rule.Match.ProcessNameType == "" {
		rule.Match.ProcessNameType = MatchTypeContains
	}
	if rule.Match.ParentName != "" && rule.Match.ParentNameType == "" {
		rule.Match.ParentNameType = MatchTypeContains
	}
}

func (e *Engine) hasExecCriteria(rule *Rule) bool {
	return rule.Match.ProcessName != "" || rule.Match.ParentName != "" ||
		rule.Match.PID != 0 || rule.Match.PPID != 0
}

func (e *Engine) indexExecRule(rule *Rule) {
	indexed := false

	// Try to index by exact process name
	if rule.Match.ProcessName != "" && rule.Match.ProcessNameType == MatchTypeExact {
		e.exactProcessNameRules[rule.Match.ProcessName] = append(
			e.exactProcessNameRules[rule.Match.ProcessName], rule)
		indexed = true
	}

	// Try to index by exact parent name
	if rule.Match.ParentName != "" && rule.Match.ParentNameType == MatchTypeExact {
		e.exactParentNameRules[rule.Match.ParentName] = append(
			e.exactParentNameRules[rule.Match.ParentName], rule)
		indexed = true
	}

	// Add to partial match list if not fully indexed or has partial matches
	needsPartialMatch := !indexed ||
		rule.Match.ProcessNameType == MatchTypeContains ||
		rule.Match.ParentNameType == MatchTypeContains

	if needsPartialMatch {
		e.partialMatchRules = append(e.partialMatchRules, rule)
	}
}

func (e *Engine) Match(event events.ProcessedEvent) []Alert {
	var alerts []Alert
	checked := make(map[*Rule]bool)

	checkRules := func(rulesToCheck []*Rule) {
		for _, rule := range rulesToCheck {
			if !checked[rule] && e.matchRule(*rule, event) {
				alerts = append(alerts, Alert{
					Rule:    *rule,
					Event:   event,
					Message: rule.Description,
				})
				checked[rule] = true
			}
		}
	}

	if rules, ok := e.exactProcessNameRules[event.Process]; ok {
		checkRules(rules)
	}
	if rules, ok := e.exactParentNameRules[event.Parent]; ok {
		checkRules(rules)
	}
	checkRules(e.partialMatchRules)

	return alerts
}

func (e *Engine) matchRule(rule Rule, event events.ProcessedEvent) bool {
	match := rule.Match
	return (match.ProcessName == "" || matchString(event.Process, match.ProcessName, match.ProcessNameType)) &&
		(match.ParentName == "" || matchString(event.Parent, match.ParentName, match.ParentNameType)) &&
		(match.PID == 0 || event.Event.PID == match.PID) &&
		(match.PPID == 0 || event.Event.PPID == match.PPID) &&
		(!match.InContainer || event.Event.CgroupID != 1)
}

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

func (e *Engine) MatchFile(filename string, pid uint32, cgroupID uint64) (bool, *Rule) {
	if rules, ok := e.exactFilenameRules[filename]; ok {
		for _, rule := range rules {
			if e.matchFileRule(*rule, filename, pid, cgroupID) {
				return true, rule
			}
		}
	}

	for _, rule := range e.filepathRules {
		if e.matchFileRule(*rule, filename, pid, cgroupID) {
			return true, rule
		}
	}

	return false, nil
}

func (e *Engine) matchFileRule(rule Rule, filename string, pid uint32, cgroupID uint64) bool {
	match := rule.Match
	return (match.Filename != "" || match.FilePath != "") &&
		(match.Filename == "" || filename == match.Filename) &&
		(match.FilePath == "" || strings.HasPrefix(filename, match.FilePath)) &&
		(!match.InContainer || cgroupID != 1) &&
		(match.PID == 0 || pid == match.PID)
}

func (e *Engine) MatchConnect(event *events.ConnectEvent) (bool, *Rule) {
	for i := range e.rules {
		rule := &e.rules[i]
		if e.matchConnectRule(*rule, event) {
			return true, rule
		}
	}
	return false, nil
}

func (e *Engine) matchConnectRule(rule Rule, event *events.ConnectEvent) bool {
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

func matchIP(eventIP, ruleIP string) bool {
	_, ipNet, err := net.ParseCIDR(ruleIP)
	if err == nil {
		ip := net.ParseIP(eventIP)
		return ip != nil && ipNet.Contains(ip)
	}
	return eventIP == ruleIP
}
