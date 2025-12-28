package rules

import (
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"aegis/pkg/events"
	"aegis/pkg/utils"
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

type InodeKey struct {
	Ino uint64
	Dev uint64
}

type RuleState string

const (
	RuleStateDraft      RuleState = "draft"      // Created but not deployed
	RuleStateTesting    RuleState = "testing"    // Deployed in testing mode
	RuleStateProduction RuleState = "production" // Deployed in production
	RuleStateArchived   RuleState = "archived"   // Disabled/removed
)

type Rule struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description" yaml:"description"`
	Severity    string         `json:"severity" yaml:"severity"`
	Match       MatchCondition `json:"match" yaml:"match"`
	Action      ActionType     `json:"action" yaml:"action"`
	Type        RuleType       `json:"type,omitempty" yaml:"type,omitempty"`

	// Lifecycle state
	State      RuleState  `json:"state" yaml:"state,omitempty"`
	CreatedAt  time.Time  `json:"created_at" yaml:"-"`
	DeployedAt *time.Time `json:"deployed_at,omitempty" yaml:"-"`
	PromotedAt *time.Time `json:"promoted_at,omitempty" yaml:"-"`

	// NEW: Validation metrics
	ActualTestingHits int      `json:"actual_testing_hits,omitempty" yaml:"-"`
	PromotionScore    float64  `json:"promotion_score,omitempty" yaml:"-"` // 0-1
	PromotionReasons  []string `json:"promotion_reasons,omitempty" yaml:"-"`

	// NEW: Metadata
	LastReviewedAt *time.Time `json:"last_reviewed_at,omitempty" yaml:"-"`
	ReviewNotes    string     `json:"review_notes,omitempty" yaml:"-"`
}

// Helper functions for rule state checks
func (r *Rule) IsTesting() bool {
	return r.State == RuleStateTesting
}

func (r *Rule) IsProduction() bool {
	return r.State == RuleStateProduction
}

func (r *Rule) IsDraft() bool {
	return r.State == RuleStateDraft || r.State == ""
}

func (r *Rule) IsActive() bool {
	return r.State == RuleStateTesting || r.State == RuleStateProduction
}

func (r *Rule) DeriveType() RuleType {
	if r.Type != "" {
		return r.Type
	}
	// Check filename first (before path keys which require Prepare())
	if r.Match.Filename != "" {
		return RuleTypeFile
	}
	if len(r.Match.ExactPathKeys()) > 0 || len(r.Match.PrefixPathKeys()) > 0 {
		return RuleTypeFile
	}
	if r.Match.DestPort != 0 || r.Match.DestIP != "" {
		return RuleTypeConnect
	}
	return RuleTypeExec
}

type MatchCondition struct {
	ProcessName     string     `yaml:"process_name,omitempty"`
	ProcessNameType MatchType  `yaml:"process_name_type,omitempty"`
	ParentName      string     `yaml:"parent_name,omitempty"`
	ParentNameType  MatchType  `yaml:"parent_name_type,omitempty"`
	PID             uint32     `yaml:"pid,omitempty"`
	PPID            uint32     `yaml:"ppid,omitempty"`
	CgroupID        string     `yaml:"cgroup_id,omitempty"`
	Filename        string     `yaml:"filename,omitempty"`
	DestPort        uint16     `yaml:"dest_port,omitempty"`
	DestIP          string     `yaml:"dest_ip,omitempty"`
	destIPNet       *net.IPNet `yaml:"-"`
	destIPPrepared  bool       `yaml:"-"`
	inode           InodeKey   `yaml:"-"`
	inodeResolved   bool       `yaml:"-"`
	pathExactKeys   []string   `yaml:"-"`
	pathPrefixKeys  []string   `yaml:"-"`
}

type RuleSet struct {
	Rules []Rule `yaml:"rules"`
}

type MatchedAlert struct {
	Rule    Rule
	Event   events.ProcessedEvent
	Message string
}

func (m *MatchCondition) Prepare() {
	if m == nil {
		return
	}

	if m.Filename != "" {
		m.prepareFilenameKeys(m.Filename)
		m.prepareInode()
	} else {
		m.pathExactKeys = nil
		m.pathPrefixKeys = nil
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

func (m *MatchCondition) MatchIP(eventIP string) bool {
	if m == nil || m.DestIP == "" {
		return true
	}
	if m.destIPNet != nil {
		ip := net.ParseIP(eventIP)
		return ip != nil && m.destIPNet.Contains(ip)
	}
	return eventIP == m.DestIP
}

func (m *MatchCondition) InodeKey() (InodeKey, bool) {
	if m == nil || !m.inodeResolved {
		return InodeKey{}, false
	}
	return m.inode, true
}

func (m *MatchCondition) ExactPathKeys() []string {
	if m == nil {
		return nil
	}
	return m.pathExactKeys
}

func (m *MatchCondition) PrefixPathKeys() []string {
	if m == nil {
		return nil
	}
	return m.pathPrefixKeys
}

func (m *MatchCondition) prepareInode() {
	if m.inodeResolved {
		return
	}

	path := m.Filename
	if path == "" {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		log.Printf("Skipping file rule for %s: %v", path, err)
		return
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		log.Printf("Skipping file rule for %s: unsupported stat type", path)
		return
	}

	m.inode = InodeKey{
		Ino: stat.Ino,
		Dev: uint64(stat.Dev),
	}
	m.inodeResolved = true
}

func (m *MatchCondition) prepareFilenameKeys(raw string) {
	path := strings.TrimSpace(raw)
	if path == "" {
		m.pathExactKeys = nil
		m.pathPrefixKeys = nil
		return
	}

	if strings.HasSuffix(path, "*") {
		base := strings.TrimSuffix(path, "*")
		base = strings.TrimSuffix(base, "/")
		variants := utils.PathVariants(base)
		m.pathExactKeys = nil
		m.pathPrefixKeys = normalizePrefixVariants(variants)
		return
	}

	variants := utils.PathVariants(path)
	m.pathExactKeys = variants
	m.pathPrefixKeys = nil
}

func normalizePrefixVariants(variants []string) []string {
	if len(variants) == 0 {
		return nil
	}
	out := make([]string, 0, len(variants))
	for _, v := range variants {
		if v == "" {
			continue
		}
		prefix := v
		if prefix != "/" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		out = append(out, prefix)
	}
	return out
}
