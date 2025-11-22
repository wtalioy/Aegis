package rules

import (
	"eulerguard/pkg/events"
	"strings"
)

type Engine struct {
	rules []Rule
}

func NewEngine(rules []Rule) *Engine {
	return &Engine{
		rules: rules,
	}
}

// Match checks if an event matches any rules and returns alerts
func (e *Engine) Match(event events.ProcessedEvent) []Alert {
	var alerts []Alert

	for _, rule := range e.rules {
		if e.matchRule(rule, event) {
			alert := Alert{
				Rule:    rule,
				Event:   event,
				Message: rule.Description,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// matchRule checks if a single rule matches the event
func (e *Engine) matchRule(rule Rule, event events.ProcessedEvent) bool {
	match := rule.Match

	if match.ProcessName != "" {
		processName := strings.TrimSpace(event.Process)
		if !strings.Contains(processName, match.ProcessName) {
			return false
		}
	}

	if match.ParentName != "" {
		parentName := strings.TrimSpace(event.Parent)
		if !strings.Contains(parentName, match.ParentName) {
			return false
		}
	}

	if match.PID != 0 {
		if event.Event.PID != match.PID {
			return false
		}
	}

	if match.PPID != 0 {
		if event.Event.PPID != match.PPID {
			return false
		}
	}

	if match.InContainer {
		if event.Event.CgroupID == 1 {
			return false
		}
	}

	return true
}

func (e *Engine) GetRules() []Rule {
	return e.rules
}

func (e *Engine) RuleCount() int {
	return len(e.rules)
}
