package rules

import "strings"

type fileEvent struct {
	filename string
	pid      uint32
	cgroupID uint64
}

type fileMatcher struct {
	exactFilenameRules map[string][]*Rule
	filepathRules      []*Rule
}

func newFileMatcher(rules []Rule) *fileMatcher {
	matcher := &fileMatcher{
		exactFilenameRules: make(map[string][]*Rule),
		filepathRules:      make([]*Rule, 0),
	}
	for i := range rules {
		rule := &rules[i]
		if rule.Match.Filename != "" {
			matcher.exactFilenameRules[rule.Match.Filename] = append(
				matcher.exactFilenameRules[rule.Match.Filename], rule)
		}
		if rule.Match.FilePath != "" {
			matcher.filepathRules = append(matcher.filepathRules, rule)
		}
	}
	return matcher
}

func (m *fileMatcher) Match(filename string, pid uint32, cgroupID uint64) (matched bool, rule *Rule, allowed bool) {
	event := fileEvent{filename: filename, pid: pid, cgroupID: cgroupID}
	return filterRulesByAction(m.getCandidateRules(filename), m.matchRule, event)
}

func (m *fileMatcher) getCandidateRules(filename string) []*Rule {
	var candidates []*Rule
	if rules, ok := m.exactFilenameRules[filename]; ok {
		candidates = append(candidates, rules...)
	}
	candidates = append(candidates, m.filepathRules...)
	return candidates
}

func (m *fileMatcher) matchRule(rule *Rule, event fileEvent) bool {
	match := rule.Match
	if match.Filename == "" && match.FilePath == "" {
		return false
	}
	if match.Filename != "" && event.filename != match.Filename {
		return false
	}
	if match.FilePath != "" && !strings.HasPrefix(event.filename, match.FilePath) {
		return false
	}
	return matchCgroupID(match.CgroupID, event.cgroupID) && matchPID(match.PID, event.pid)
}
