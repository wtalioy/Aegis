package profiler

import (
	"fmt"
	"os"
	"sync"

	"eulerguard/pkg/events"
	"eulerguard/pkg/types"
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
		File:     utils.SimplifyPath(filename),
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

func (p *Profiler) Counts() (exec, file, connect int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for profile := range p.profiles {
		switch profile.Type {
		case events.EventTypeExec:
			exec++
		case events.EventTypeFileOpen:
			file++
		case events.EventTypeConnect:
			connect++
		}
	}
	return
}

func (p *Profiler) GetProfiles() []BehaviorProfile {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]BehaviorProfile, 0, len(p.profiles))
	for profile := range p.profiles {
		result = append(result, profile)
	}
	return result
}

func (p *Profiler) GenerateRules() []types.Rule {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ruleList := make([]types.Rule, 0, len(p.profiles))

	for profile := range p.profiles {
		rule := p.profileToRule(profile)
		ruleList = append(ruleList, rule)
	}

	return ruleList
}

func (p *Profiler) GenerateRulesFiltered(indices []int) []types.Rule {
	allRules := p.GenerateRules()
	if len(indices) == 0 {
		return allRules
	}

	indexSet := make(map[int]bool)
	for _, i := range indices {
		indexSet[i] = true
	}

	result := make([]types.Rule, 0, len(indices))
	for i, rule := range allRules {
		if indexSet[i] {
			result = append(result, rule)
		}
	}
	return result
}

func (p *Profiler) profileToRule(profile BehaviorProfile) types.Rule {
	rule := types.Rule{
		Description: "Auto-generated from learning mode",
		Severity:    "info",
		Action:      "allow",
	}

	switch profile.Type {
	case events.EventTypeExec:
		rule.Name = fmt.Sprintf("Allow %s from %s", profile.Process, profile.Parent)
		rule.Type = types.RuleTypeExec
		rule.Match = types.MatchCondition{
			ProcessName:     profile.Process,
			ProcessNameType: types.MatchTypeExact,
			ParentName:      profile.Parent,
			ParentNameType:  types.MatchTypeExact,
		}

	case events.EventTypeFileOpen:
		rule.Name = fmt.Sprintf("Allow access to %s", profile.File)
		rule.Type = types.RuleTypeFile
		rule.Match = types.MatchCondition{
			Filename: profile.File,
		}

	case events.EventTypeConnect:
		rule.Name = fmt.Sprintf("Allow connection to port %d", profile.Port)
		rule.Type = types.RuleTypeConnect
		rule.Match = types.MatchCondition{
			DestPort: profile.Port,
		}
	}

	return rule
}

func (p *Profiler) SaveYAML(path string) error {
	ruleList := p.GenerateRules()

	ruleSet := types.RuleSet{
		Rules: ruleList,
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create rules file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(ruleSet); err != nil {
		return fmt.Errorf("failed to encode rules to YAML: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close encoder: %w", err)
	}

	return nil
}
