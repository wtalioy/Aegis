package rules

import "strings"

type fileMatcher struct {
	exactFilenameRules map[string][]*Rule
	filepathRules      []*Rule
}

func newFileMatcher(rules []Rule) *fileMatcher {
	matcher := &fileMatcher{
		exactFilenameRules: make(map[string][]*Rule),
		filepathRules:      make([]*Rule, 0),
	}
	matcher.indexRules(rules)
	return matcher
}

func (m *fileMatcher) indexRules(rules []Rule) {
	for i := range rules {
		rule := &rules[i]
		if rule.Match.Filename != "" {
			m.exactFilenameRules[rule.Match.Filename] = append(
				m.exactFilenameRules[rule.Match.Filename], rule)
		}
		if rule.Match.FilePath != "" {
			m.filepathRules = append(m.filepathRules, rule)
		}
	}
}

func (m *fileMatcher) Match(filename string, pid uint32, cgroupID uint64) (bool, *Rule) {
	if match := m.matchFromExact(filename, pid, cgroupID); match != nil {
		return true, match
	}
	if match := m.matchFromPaths(filename, pid, cgroupID); match != nil {
		return true, match
	}
	return false, nil
}

func (m *fileMatcher) matchFromExact(filename string, pid uint32, cgroupID uint64) *Rule {
	if rules, ok := m.exactFilenameRules[filename]; ok {
		return m.findMatch(rules, filename, pid, cgroupID)
	}
	return nil
}

func (m *fileMatcher) matchFromPaths(filename string, pid uint32, cgroupID uint64) *Rule {
	return m.findMatch(m.filepathRules, filename, pid, cgroupID)
}

func (m *fileMatcher) findMatch(rules []*Rule, filename string, pid uint32, cgroupID uint64) *Rule {
	for _, rule := range rules {
		if m.matchRule(rule, filename, pid, cgroupID) {
			return rule
		}
	}
	return nil
}

func (m *fileMatcher) matchRule(rule *Rule, filename string, pid uint32, cgroupID uint64) bool {
	match := rule.Match
	if match.Filename == "" && match.FilePath == "" {
		return false
	}
	if match.Filename != "" && filename != match.Filename {
		return false
	}
	if match.FilePath != "" && !strings.HasPrefix(filename, match.FilePath) {
		return false
	}
	if match.InContainer && cgroupID == 1 {
		return false
	}
	if match.PID != 0 && pid != match.PID {
		return false
	}
	return true
}
