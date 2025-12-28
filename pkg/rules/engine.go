package rules

import (
	"aegis/pkg/events"
)

type Engine struct {
	rules          []Rule
	execMatcher    *execMatcher
	fileMatcher    *fileMatcher
	connectMatcher *connectMatcher
	testingBuffer  *TestingBuffer
}

func NewEngine(rules []Rule) *Engine {
	// Separate active rules (testing/production) from draft rules.
	// Draft rules should not be included in matchers.
	var activeRules []Rule
	for i := range rules {
		rules[i].Match.Prepare()
		// Only include rules that are active (testing or production)
		// Draft rules and empty state rules are excluded from matching
		if rules[i].IsActive() {
			activeRules = append(activeRules, rules[i])
		}
	}
	return &Engine{
		rules:          rules, // Keep all rules for GetRules(), but only active ones in matchers
		execMatcher:    newExecMatcher(activeRules),
		fileMatcher:    newFileMatcher(activeRules),
		connectMatcher: newConnectMatcher(activeRules),
		testingBuffer:  NewTestingBuffer(10000),
	}
}

func (e *Engine) MatchExec(event events.ProcessedEvent) (matched bool, rule *Rule, allowed bool) {
	if e.execMatcher == nil {
		return false, nil, false
	}
	return e.execMatcher.Match(event)
}

func (e *Engine) CollectExecAlerts(event events.ProcessedEvent) []MatchedAlert {
	if e.execMatcher == nil {
		return nil
	}
	return e.execMatcher.CollectAlerts(event)
}

func (e *Engine) MatchFile(ino, dev uint64, filename string, pid uint32, cgroupID uint64) (matched bool, rule *Rule, allowed bool) {
	if e.fileMatcher == nil {
		return false, nil, false
	}
	return e.fileMatcher.Match(ino, dev, filename, pid, cgroupID)
}

func (e *Engine) MatchConnect(event *events.ConnectEvent) (matched bool, rule *Rule, allowed bool) {
	if e.connectMatcher == nil {
		return false, nil, false
	}
	return e.connectMatcher.Match(event)
}

func (e *Engine) GetRules() []Rule {
	return e.rules
}

func (e *Engine) GetTestingBuffer() *TestingBuffer {
	return e.testingBuffer
}
