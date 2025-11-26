package rules

import (
	"eulerguard/pkg/events"
)

type MatchType string

type ActionType string

const (
	ActionAllow ActionType = "allow"
	ActionAlert ActionType = "alert"
)

const (
	MatchTypeExact    MatchType = "exact"
	MatchTypeContains MatchType = "contains"
	MatchTypePrefix   MatchType = "prefix"
)

type RuleType string

const (
	RuleTypeExec    RuleType = "exec"
	RuleTypeFile    RuleType = "file"
	RuleTypeConnect RuleType = "connect"
)

type Rule struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Severity    string         `yaml:"severity"`
	Match       MatchCondition `yaml:"match"`
	Action      ActionType     `yaml:"action"`
	Type        RuleType       `yaml:"type,omitempty"`
}

func (r *Rule) DeriveType() RuleType {
	if r.Type != "" {
		return r.Type
	}
	if r.Match.Filename != "" || r.Match.FilePath != "" {
		return RuleTypeFile
	}
	if r.Match.DestPort != 0 || r.Match.DestIP != "" {
		return RuleTypeConnect
	}
	return RuleTypeExec
}

type MatchCondition struct {
	ProcessName     string    `yaml:"process_name,omitempty"`
	ProcessNameType MatchType `yaml:"process_name_type,omitempty"`
	ParentName      string    `yaml:"parent_name,omitempty"`
	ParentNameType  MatchType `yaml:"parent_name_type,omitempty"`
	PID             uint32    `yaml:"pid,omitempty"`
	PPID            uint32    `yaml:"ppid,omitempty"`
	CgroupID        string    `yaml:"cgroup_id,omitempty"`
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
