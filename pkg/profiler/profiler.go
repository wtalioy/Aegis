package profiler

import (
	"fmt"
	"os"
	"sync"

	"eulerguard/pkg/events"
	"eulerguard/pkg/rules"
	"eulerguard/pkg/utils"

	"gopkg.in/yaml.v3"
)

type BehaviorProfile struct {
	Type     events.EventType
	Process  string
	Parent   string
	File     string
	Port     uint16
	CgroupID uint64
}

type Profiler struct {
	mu       sync.RWMutex
	profiles map[BehaviorProfile]struct{}
	active   bool
}

var _ events.EventHandler = (*Profiler)(nil)

func NewProfiler() *Profiler {
	return &Profiler{
		profiles: make(map[BehaviorProfile]struct{}),
		active:   true,
	}
}

func (p *Profiler) IsActive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.active
}

func (p *Profiler) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.active = false
}

func (p *Profiler) HandleExec(ev events.ExecEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return
	}

	profile := BehaviorProfile{
		Type:     events.EventTypeExec,
		Process:  utils.ExtractCString(ev.Comm[:]),
		Parent:   utils.ExtractCString(ev.PComm[:]),
		CgroupID: ev.CgroupID,
	}

	p.profiles[profile] = struct{}{}
}

func (p *Profiler) HandleFileOpen(ev events.FileOpenEvent, filename string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return
	}

	profile := BehaviorProfile{
		Type:     events.EventTypeFileOpen,
		File:     filename,
		CgroupID: ev.CgroupID,
	}

	p.profiles[profile] = struct{}{}
}

func (p *Profiler) HandleConnect(ev events.ConnectEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return
	}

	profile := BehaviorProfile{
		Type:     events.EventTypeConnect,
		Port:     ev.Port,
		CgroupID: ev.CgroupID,
	}

	p.profiles[profile] = struct{}{}
}

func (p *Profiler) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.profiles)
}

func (p *Profiler) GenerateRules() []rules.Rule {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ruleList := make([]rules.Rule, 0, len(p.profiles))

	for profile := range p.profiles {
		rule := p.profileToRule(profile)
		ruleList = append(ruleList, rule)
	}

	return ruleList
}

func (p *Profiler) profileToRule(profile BehaviorProfile) rules.Rule {
	rule := rules.Rule{
		Description: "Auto-generated from learning mode",
		Severity:    "info",
		Action:      "allow",
	}

	switch profile.Type {
	case events.EventTypeExec:
		rule.Name = fmt.Sprintf("Allow %s from %s", profile.Process, profile.Parent)
		rule.Match = rules.MatchCondition{
			ProcessName:     profile.Process,
			ProcessNameType: rules.MatchTypeExact,
			ParentName:      profile.Parent,
			ParentNameType:  rules.MatchTypeExact,
		}
		if profile.CgroupID > 1 {
			rule.Match.InContainer = true
		}

	case events.EventTypeFileOpen:
		rule.Name = fmt.Sprintf("Allow access to %s", profile.File)
		rule.Match = rules.MatchCondition{
			Filename: profile.File,
		}
		if profile.CgroupID > 1 {
			rule.Match.InContainer = true
		}

	case events.EventTypeConnect:
		rule.Name = fmt.Sprintf("Allow connection to port %d", profile.Port)
		rule.Match = rules.MatchCondition{
			DestPort: profile.Port,
		}
		if profile.CgroupID > 1 {
			rule.Match.InContainer = true
		}
	}

	return rule
}

func (p *Profiler) SaveYAML(path string) error {
	ruleList := p.GenerateRules()

	ruleSet := rules.RuleSet{
		Rules: ruleList,
	}

	data, err := yaml.Marshal(ruleSet)
	if err != nil {
		return fmt.Errorf("failed to marshal rules to YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write rules file: %w", err)
	}

	return nil
}
