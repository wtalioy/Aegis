package rules

import "eulerguard/pkg/events"

type Engine struct {
	rules []Rule

	execMatcher    *execMatcher
	fileMatcher    *fileMatcher
	connectMatcher *connectMatcher
}

func NewEngine(rules []Rule) *Engine {
	engine := &Engine{
		rules: rules,
	}
	engine.execMatcher = newExecMatcher(engine.rules)
	engine.fileMatcher = newFileMatcher(engine.rules)
	engine.connectMatcher = newConnectMatcher(engine.rules)
	return engine
}

func (e *Engine) Match(event events.ProcessedEvent) []Alert {
	if e.execMatcher == nil {
		return nil
	}
	return e.execMatcher.Match(event)
}

func (e *Engine) MatchFile(filename string, pid uint32, cgroupID uint64) (bool, *Rule) {
	if e.fileMatcher == nil {
		return false, nil
	}
	return e.fileMatcher.Match(filename, pid, cgroupID)
}

func (e *Engine) MatchConnect(event *events.ConnectEvent) (bool, *Rule) {
	if e.connectMatcher == nil {
		return false, nil
	}
	return e.connectMatcher.Match(event)
}
