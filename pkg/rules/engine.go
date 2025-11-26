package rules

import "eulerguard/pkg/events"

type Engine struct {
	rules          []Rule
	execMatcher    *execMatcher
	fileMatcher    *fileMatcher
	connectMatcher *connectMatcher
}

func NewEngine(rules []Rule) *Engine {
	return &Engine{
		rules:          rules,
		execMatcher:    newExecMatcher(rules),
		fileMatcher:    newFileMatcher(rules),
		connectMatcher: newConnectMatcher(rules),
	}
}

func (e *Engine) MatchExec(event events.ProcessedEvent) (matched bool, rule *Rule, allowed bool) {
	if e.execMatcher == nil {
		return false, nil, false
	}
	return e.execMatcher.Match(event)
}

func (e *Engine) CollectExecAlerts(event events.ProcessedEvent) []Alert {
	if e.execMatcher == nil {
		return nil
	}
	return e.execMatcher.CollectAlerts(event)
}

func (e *Engine) MatchFile(filename string, pid uint32, cgroupID uint64) (matched bool, rule *Rule, allowed bool) {
	if e.fileMatcher == nil {
		return false, nil, false
	}
	return e.fileMatcher.Match(filename, pid, cgroupID)
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
