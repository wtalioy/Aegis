package rules

import (
	"eulerguard/pkg/events"
)

type MatchType string

const (
	MatchTypeExact    MatchType = "exact"
	MatchTypeContains MatchType = "contains"
	MatchTypePrefix   MatchType = "prefix"
)

type Rule struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Severity    string         `yaml:"severity"`
	Match       MatchCondition `yaml:"match"`
	Action      string         `yaml:"action"`
}

type MatchCondition struct {
	ProcessName     string    `yaml:"process_name,omitempty"`
	ProcessNameType MatchType `yaml:"process_name_type,omitempty"`
	ParentName      string    `yaml:"parent_name,omitempty"`
	ParentNameType  MatchType `yaml:"parent_name_type,omitempty"`
	PID             uint32    `yaml:"pid,omitempty"`
	PPID            uint32    `yaml:"ppid,omitempty"`
	InContainer     bool      `yaml:"in_container,omitempty"`
	Filename        string    `yaml:"filename,omitempty"`
	FilePath        string    `yaml:"file_path,omitempty"`
	DestPort        uint16    `yaml:"dest_port,omitempty"`
	DestIP          string    `yaml:"dest_ip,omitempty"`
}

type RuleSet struct {
	Rules []Rule `yaml:"rules"`
}

type Alert struct {
	Rule    Rule
	Event   events.ProcessedEvent
	Message string
}
