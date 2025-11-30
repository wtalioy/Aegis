package rules

import (
	"net"

	"eulerguard/pkg/events"
)

type MatchType string

type ActionType string

const (
	ActionAllow ActionType = "allow"
	ActionAlert ActionType = "alert"
	ActionBlock ActionType = "block"
)

const (
	BPFActionMonitor uint8 = 1
	BPFActionBlock   uint8 = 2
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
	ProcessName        string     `yaml:"process_name,omitempty"`
	ProcessNameType    MatchType  `yaml:"process_name_type,omitempty"`
	ParentName         string     `yaml:"parent_name,omitempty"`
	ParentNameType     MatchType  `yaml:"parent_name_type,omitempty"`
	PID                uint32     `yaml:"pid,omitempty"`
	PPID               uint32     `yaml:"ppid,omitempty"`
	CgroupID           string     `yaml:"cgroup_id,omitempty"`
	Filename           string     `yaml:"filename,omitempty"`
	FilePath           string     `yaml:"file_path,omitempty"`
	DestPort           uint16     `yaml:"dest_port,omitempty"`
	DestIP             string     `yaml:"dest_ip,omitempty"`
	normalizedFilename string     `yaml:"-"`
	destIPNet          *net.IPNet `yaml:"-"`
	destIPPrepared     bool       `yaml:"-"`
}

type RuleSet struct {
	Rules []Rule `yaml:"rules"`
}

type Alert struct {
	Rule    Rule
	Event   events.ProcessedEvent
	Message string
}

func (m *MatchCondition) prepare() {
	if m == nil {
		return
	}

	if m.Filename != "" {
		m.normalizedFilename = normalizeFilename(m.Filename)
	} else {
		m.normalizedFilename = ""
	}

	if !m.destIPPrepared {
		m.destIPPrepared = true
		if m.DestIP == "" {
			m.destIPNet = nil
			return
		}
		if _, cidr, err := net.ParseCIDR(m.DestIP); err == nil {
			m.destIPNet = cidr
		} else {
			m.destIPNet = nil
		}
	}
}

func (m *MatchCondition) matchIP(eventIP string) bool {
	if m == nil || m.DestIP == "" {
		return true
	}
	if m.destIPNet != nil {
		ip := net.ParseIP(eventIP)
		return ip != nil && m.destIPNet.Contains(ip)
	}
	return eventIP == m.DestIP
}
