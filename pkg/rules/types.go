package rules

import (
	"eulerguard/pkg/events"
)

type Rule struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Severity    string         `yaml:"severity"`
	Match       MatchCondition `yaml:"match"`
	Action      string         `yaml:"action"`
}

type MatchCondition struct {
	ProcessName string `yaml:"process_name,omitempty"`
	ParentName  string `yaml:"parent_name,omitempty"`
	PID         uint32 `yaml:"pid,omitempty"`
	PPID        uint32 `yaml:"ppid,omitempty"`
	InContainer bool   `yaml:"in_container,omitempty"`
}

type RuleSet struct {
	Rules []Rule `yaml:"rules"`
}

type Alert struct {
	Rule    Rule
	Event   events.ProcessedEvent
	Message string
}
