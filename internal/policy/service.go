package policy

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"aegis/internal/platform/events"
	"aegis/internal/platform/storage"
	"aegis/internal/policy/rules"
	"aegis/internal/system"
	"aegis/internal/telemetry"
)

type RuleRepository interface {
	Load() ([]Rule, error)
	Save([]Rule) error
}

type KernelSync interface {
	SyncRules([]Rule) error
}

type DecisionType string

const (
	DecisionNoMatch    DecisionType = "no_match"
	DecisionAllow      DecisionType = "allow"
	DecisionAlert      DecisionType = "alert"
	DecisionBlock      DecisionType = "block"
	DecisionTestingHit DecisionType = "testing_hit"
)

type Decision struct {
	Type   DecisionType
	Rule   *Rule
	Alerts []system.Alert
}

type Service struct {
	mu             sync.RWMutex
	repo           RuleRepository
	kernelSync     KernelSync
	ruleList       []Rule
	engine         *rules.Engine
	validation     *rules.ValidationService
	observationMin int
	minHits        int
}

func NewService(repo RuleRepository, kernelSync KernelSync, observationMin int, minHits int) *Service {
	return &Service{
		repo:           repo,
		kernelSync:     kernelSync,
		observationMin: observationMin,
		minHits:        minHits,
	}
}

func (s *Service) SetKernelSync(kernelSync KernelSync) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.kernelSync = kernelSync
}

func (s *Service) UpdateThresholds(observationMin, minHits int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.observationMin = observationMin
	s.minHits = minHits
	if s.engine != nil {
		s.validation = rules.NewValidationService(s.engine.GetTestingBuffer(), s.observationMin, s.minHits)
	}
}

func (s *Service) Bootstrap(ruleList []Rule) error {
	return s.replaceRules(ruleList)
}

func (s *Service) Load() error {
	ruleList, err := s.repo.Load()
	if err != nil {
		return err
	}
	return s.replaceRules(ruleList)
}

func (s *Service) Reload() error {
	return s.Load()
}

func (s *Service) List() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Rule, len(s.ruleList))
	copy(result, s.ruleList)
	return result
}

func (s *Service) Get(name string) (*Rule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.ruleList {
		if s.ruleList[i].Name == name {
			rule := s.ruleList[i]
			return &rule, true
		}
	}
	return nil, false
}

func (s *Service) Engine() *rules.Engine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine
}

func (s *Service) Create(rule Rule) (Rule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	if rule.State == "" {
		rule.State = rules.RuleStateTesting
	}
	if (rule.State == rules.RuleStateTesting || rule.State == rules.RuleStateProduction) && rule.DeployedAt == nil {
		rule.DeployedAt = &now
	}

	next := append(append([]rules.Rule(nil), s.ruleList...), rule)
	if err := s.saveAndReplaceLocked(next); err != nil {
		return Rule{}, err
	}
	return rule, nil
}

func (s *Service) Update(name string, update Rule) (Rule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := append([]rules.Rule(nil), s.ruleList...)
	for i := range next {
		if next[i].Name != name {
			continue
		}
		update.Name = name
		if update.CreatedAt.IsZero() {
			update.CreatedAt = next[i].CreatedAt
		}
		if update.DeployedAt == nil {
			update.DeployedAt = next[i].DeployedAt
		}
		if update.PromotedAt == nil {
			update.PromotedAt = next[i].PromotedAt
		}
		if update.State == "" {
			update.State = next[i].State
		}
		next[i] = update
		if err := s.saveAndReplaceLocked(next); err != nil {
			return Rule{}, err
		}
		return update, nil
	}
	return Rule{}, fmt.Errorf("rule %s not found", name)
}

func (s *Service) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := make([]rules.Rule, 0, len(s.ruleList))
	found := false
	for _, rule := range s.ruleList {
		if rule.Name == name {
			found = true
			continue
		}
		next = append(next, rule)
	}
	if !found {
		return fmt.Errorf("rule %s not found", name)
	}
	return s.saveAndReplaceLocked(next)
}

func (s *Service) Promote(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := append([]rules.Rule(nil), s.ruleList...)
	for i := range next {
		if next[i].Name != name {
			continue
		}
		now := time.Now()
		next[i].State = rules.RuleStateProduction
		next[i].PromotedAt = &now
		if next[i].DeployedAt == nil {
			next[i].DeployedAt = &now
		}
		return s.saveAndReplaceLocked(next)
	}
	return fmt.Errorf("rule %s not found", name)
}

func (s *Service) Validation(name string) (PromotionReadiness, TestingStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.engine == nil || s.validation == nil {
		return PromotionReadiness{}, TestingStats{}, fmt.Errorf("rule engine not available")
	}
	for i := range s.ruleList {
		if s.ruleList[i].Name == name {
			return s.validation.CalculatePromotionReadiness(&s.ruleList[i]), s.engine.GetTestingBuffer().GetStats(name), nil
		}
	}
	return PromotionReadiness{}, TestingStats{}, fmt.Errorf("rule %s not found", name)
}

func (s *Service) TestingRules() []TestingRuleStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]TestingRuleStatus, 0)
	if s.engine == nil || s.validation == nil {
		return result
	}

	for _, rule := range s.ruleList {
		if !rule.IsTesting() {
			continue
		}
		result = append(result, TestingRuleStatus{
			Rule:       rule,
			Validation: s.validation.CalculatePromotionReadiness(&rule),
			Stats:      s.engine.GetTestingBuffer().GetStats(rule.Name),
		})
	}

	return result
}

func (s *Service) Evaluate(record *telemetry.Record) Decision {
	s.mu.RLock()
	engine := s.engine
	s.mu.RUnlock()
	if engine == nil || record == nil {
		return Decision{Type: DecisionNoMatch}
	}
	event := &record.Event

	switch event.Type {
	case telemetry.EventTypeExec:
		return s.evaluateExec(engine, record)
	case telemetry.EventTypeFile:
		return s.evaluateFile(engine, record)
	case telemetry.EventTypeConnect:
		return s.evaluateConnect(engine, record)
	default:
		return Decision{Type: DecisionNoMatch}
	}
}

func (s *Service) replaceRules(ruleList []Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.replaceRulesLocked(ruleList)
}

func (s *Service) replaceRulesLocked(ruleList []Rule) error {
	engine := rules.NewEngine(ruleList)
	s.ruleList = append([]rules.Rule(nil), ruleList...)
	s.engine = engine
	s.validation = rules.NewValidationService(engine.GetTestingBuffer(), s.observationMin, s.minHits)
	if s.kernelSync != nil {
		if err := s.kernelSync.SyncRules(ruleList); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) saveAndReplaceLocked(ruleList []Rule) error {
	if err := s.repo.Save(ruleList); err != nil {
		return err
	}
	return s.replaceRulesLocked(ruleList)
}

func (s *Service) evaluateExec(engine *rules.Engine, record *telemetry.Record) Decision {
	event := &record.Event
	raw, ok := eventFromRawExec(record)
	if !ok {
		return Decision{Type: DecisionNoMatch}
	}

	processed := events.ProcessedEvent{
		Event:     raw,
		Timestamp: raw.Hdr.Timestamp(),
		Process:   event.ProcessName,
		Parent:    event.ParentName,
	}
	matched, rule, allowed := engine.MatchExec(processed)
	if allowed {
		return Decision{Type: DecisionAllow, Rule: rule}
	}

	alerts := engine.CollectExecAlerts(processed)
	if event.Blocked && len(alerts) == 0 {
		alert := system.Alert{
			ID:          alertID("exec", event.PID),
			Timestamp:   event.Timestamp.UnixMilli(),
			Severity:    "critical",
			RuleName:    "Kernel Blocked Execution",
			Description: fmt.Sprintf("Process execution blocked by kernel: %s", event.ProcessName),
			PID:         event.PID,
			ProcessName: event.ProcessName,
			ParentName:  event.ParentName,
			CgroupID:    strconv.FormatUint(event.CgroupID, 10),
			Action:      "block",
			Blocked:     true,
		}
		return Decision{Type: DecisionBlock, Alerts: []system.Alert{alert}}
	}

	if len(alerts) == 0 {
		if matched && rule != nil && rule.IsTesting() {
			return Decision{Type: DecisionTestingHit, Rule: rule}
		}
		return Decision{Type: DecisionNoMatch}
	}

	out := make([]system.Alert, 0, len(alerts))
	decisionType := DecisionAlert
	for _, alert := range alerts {
		severity := alert.Rule.Severity
		if event.Blocked && severity != "critical" {
			severity = "critical"
		}
		if alert.Rule.Action == rules.ActionBlock || event.Blocked {
			decisionType = DecisionBlock
		}
		out = append(out, system.Alert{
			ID:          alertID("exec", event.PID),
			Timestamp:   event.Timestamp.UnixMilli(),
			Severity:    severity,
			RuleName:    alert.Rule.Name,
			Description: alert.Rule.Description,
			PID:         event.PID,
			ProcessName: event.ProcessName,
			ParentName:  event.ParentName,
			CgroupID:    strconv.FormatUint(event.CgroupID, 10),
			Action:      string(alert.Rule.Action),
			Blocked:     event.Blocked,
		})
	}
	return Decision{Type: decisionType, Rule: rule, Alerts: out}
}

func (s *Service) evaluateFile(engine *rules.Engine, record *telemetry.Record) Decision {
	event := &record.Event
	raw, ok := eventFromRawFile(record)
	if !ok {
		return Decision{Type: DecisionNoMatch}
	}

	matched, rule, allowed := engine.MatchFile(raw.Ino, raw.Dev, event.Filename, raw.Hdr.PID, raw.Hdr.CgroupID)
	if event.Blocked && (!matched || rule == nil) {
		alert := system.Alert{
			ID:          alertID("file", event.PID),
			Timestamp:   event.Timestamp.UnixMilli(),
			Severity:    "critical",
			RuleName:    "Kernel Blocked File Access",
			Description: fmt.Sprintf("File access blocked by kernel: %s", event.Filename),
			PID:         event.PID,
			ProcessName: event.ProcessName,
			CgroupID:    strconv.FormatUint(event.CgroupID, 10),
			Action:      "block",
			Blocked:     true,
		}
		return Decision{Type: DecisionBlock, Alerts: []system.Alert{alert}}
	}
	if !matched || rule == nil {
		return Decision{Type: DecisionNoMatch}
	}
	if allowed {
		return Decision{Type: DecisionAllow, Rule: rule}
	}
	if rule.IsTesting() {
		if testingBuffer := engine.GetTestingBuffer(); testingBuffer != nil {
			testingBuffer.RecordHit(&rules.TestingHit{
				RuleName:    rule.Name,
				HitTime:     raw.Hdr.Timestamp(),
				EventType:   events.EventTypeFileOpen,
				EventData:   &raw,
				PID:         raw.Hdr.PID,
				ProcessName: event.ProcessName,
			})
		}
		return Decision{Type: DecisionTestingHit, Rule: rule}
	}
	severity := rule.Severity
	if event.Blocked && severity != "critical" {
		severity = "critical"
	}
	alert := system.Alert{
		ID:          alertID("file", event.PID),
		Timestamp:   event.Timestamp.UnixMilli(),
		Severity:    severity,
		RuleName:    rule.Name,
		Description: fmt.Sprintf("%s: %s", rule.Description, event.Filename),
		PID:         event.PID,
		ProcessName: event.ProcessName,
		CgroupID:    strconv.FormatUint(event.CgroupID, 10),
		Action:      string(rule.Action),
		Blocked:     event.Blocked,
	}
	alertType := DecisionAlert
	if rule.Action == rules.ActionBlock || event.Blocked {
		alertType = DecisionBlock
	}
	return Decision{Type: alertType, Rule: rule, Alerts: []system.Alert{alert}}
}

func (s *Service) evaluateConnect(engine *rules.Engine, record *telemetry.Record) Decision {
	event := &record.Event
	raw, ok := eventFromRawConnect(record)
	if !ok {
		return Decision{Type: DecisionNoMatch}
	}

	matched, rule, allowed := engine.MatchConnect(&raw)
	if event.Blocked && (!matched || rule == nil) {
		alert := system.Alert{
			ID:          alertID("net", event.PID),
			Timestamp:   event.Timestamp.UnixMilli(),
			Severity:    "critical",
			RuleName:    "Kernel Blocked Connection",
			Description: fmt.Sprintf("Network connection blocked by kernel: %s", event.Address),
			PID:         event.PID,
			ProcessName: event.ProcessName,
			CgroupID:    strconv.FormatUint(event.CgroupID, 10),
			Action:      "block",
			Blocked:     true,
		}
		return Decision{Type: DecisionBlock, Alerts: []system.Alert{alert}}
	}
	if !matched || rule == nil {
		return Decision{Type: DecisionNoMatch}
	}
	if allowed {
		return Decision{Type: DecisionAllow, Rule: rule}
	}
	if rule.IsTesting() {
		if testingBuffer := engine.GetTestingBuffer(); testingBuffer != nil {
			testingBuffer.RecordHit(&rules.TestingHit{
				RuleName:    rule.Name,
				HitTime:     raw.Hdr.Timestamp(),
				EventType:   events.EventTypeConnect,
				EventData:   &raw,
				PID:         raw.Hdr.PID,
				ProcessName: event.ProcessName,
			})
		}
		return Decision{Type: DecisionTestingHit, Rule: rule}
	}
	severity := rule.Severity
	if event.Blocked && severity != "critical" {
		severity = "critical"
	}
	alert := system.Alert{
		ID:          alertID("net", event.PID),
		Timestamp:   event.Timestamp.UnixMilli(),
		Severity:    severity,
		RuleName:    rule.Name,
		Description: rule.Description,
		PID:         event.PID,
		ProcessName: event.ProcessName,
		CgroupID:    strconv.FormatUint(event.CgroupID, 10),
		Action:      string(rule.Action),
		Blocked:     event.Blocked,
	}
	alertType := DecisionAlert
	if rule.Action == rules.ActionBlock || event.Blocked {
		alertType = DecisionBlock
	}
	return Decision{Type: alertType, Rule: rule, Alerts: []system.Alert{alert}}
}

func alertID(prefix string, pid uint32) string {
	return fmt.Sprintf("%s-%d-%d", prefix, pid, time.Now().UnixNano())
}

func eventFromRawExec(record *telemetry.Record) (events.ExecEvent, bool) {
	raw, ok := rawEvent(record)
	if !ok {
		return events.ExecEvent{}, false
	}
	switch raw := raw.Data.(type) {
	case events.ExecEvent:
		return raw, true
	case *events.ExecEvent:
		return *raw, true
	default:
		return events.ExecEvent{}, false
	}
}

func eventFromRawFile(record *telemetry.Record) (events.FileOpenEvent, bool) {
	raw, ok := rawEvent(record)
	if !ok {
		return events.FileOpenEvent{}, false
	}
	switch raw := raw.Data.(type) {
	case events.FileOpenEvent:
		return raw, true
	case *events.FileOpenEvent:
		return *raw, true
	default:
		return events.FileOpenEvent{}, false
	}
}

func eventFromRawConnect(record *telemetry.Record) (events.ConnectEvent, bool) {
	raw, ok := rawEvent(record)
	if !ok {
		return events.ConnectEvent{}, false
	}
	switch raw := raw.Data.(type) {
	case events.ConnectEvent:
		return raw, true
	case *events.ConnectEvent:
		return *raw, true
	default:
		return events.ConnectEvent{}, false
	}
}

func rawEvent(record *telemetry.Record) (*storage.Event, bool) {
	if record == nil || record.Raw == nil {
		return nil, false
	}
	return record.Raw, true
}
