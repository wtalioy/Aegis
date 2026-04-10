package analysis

import (
	"context"
	"fmt"
	"time"

	"aegis/internal/analysis/sentinel"
	"aegis/internal/analysis/types"
	aiservice "aegis/internal/platform/ai/service"
	internalconfig "aegis/internal/platform/config"
	"aegis/internal/platform/storage"
	"aegis/internal/policy"
	"aegis/internal/shared/metrics"
	"aegis/internal/telemetry"
	"aegis/internal/telemetry/proc"
	"aegis/internal/telemetry/workload"
)

type EventContextReader interface {
	RawStore() storage.EventStore
	ProcessTree() *proc.ProcessTree
	Workloads() *workload.Registry
	Profiles() *proc.ProfileRegistry
	Get(id string) (*telemetry.Record, bool)
}

type RuleReader interface {
	Engine() *policy.Engine
}

type Service struct {
	ai        *aiservice.Service
	telemetry EventContextReader
	policy    RuleReader
	stats     metrics.StatsProvider
	sentinel  *sentinel.Sentinel
}

func NewService(cfg internalconfig.Config, telemetryService EventContextReader, policyService RuleReader, stats metrics.StatsProvider) (*Service, error) {
	if cfg.Analysis.Mode == "" || cfg.Analysis.Mode == "disabled" {
		return nil, nil
	}
	aiService, err := aiservice.NewService(cfg.AIOptions())
	if err != nil {
		return nil, err
	}
	return &Service{
		ai:        aiService,
		telemetry: telemetryService,
		policy:    policyService,
		stats:     stats,
	}, nil
}

func (s *Service) Restart(cfg internalconfig.Config, telemetryService EventContextReader, policyService RuleReader, stats metrics.StatsProvider) (*Service, error) {
	if s != nil {
		s.StopSentinel()
	}
	return NewService(cfg, telemetryService, policyService, stats)
}

func (s *Service) IsEnabled() bool {
	return s != nil && s.ai != nil && s.ai.IsEnabled()
}

func (s *Service) Status() types.StatusDTO {
	if s.ai == nil {
		return types.StatusDTO{Status: "unavailable"}
	}
	return s.ai.GetStatus()
}

func (s *Service) Diagnose(ctx context.Context) (*types.DiagnosisResult, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI service not available")
	}
	return s.ai.Diagnose(ctx, s.stats, s.telemetry.Workloads(), s.telemetry.RawStore(), s.telemetry.ProcessTree())
}

func (s *Service) Chat(ctx context.Context, sessionID, message string) (*types.ChatResponse, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI service not available")
	}
	return s.ai.Chat(ctx, sessionID, message, s.stats, s.telemetry.Workloads(), s.telemetry.RawStore(), s.telemetry.ProcessTree())
}

func (s *Service) ChatStream(ctx context.Context, sessionID, message string) (<-chan types.ChatStreamToken, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI service not available")
	}
	return s.ai.ChatStream(ctx, sessionID, message, s.stats, s.telemetry.Workloads(), s.telemetry.RawStore(), s.telemetry.ProcessTree())
}

func (s *Service) ChatHistory(sessionID string) []types.Message {
	if s.ai == nil {
		return nil
	}
	return s.ai.GetChatHistory(sessionID)
}

func (s *Service) ClearChat(sessionID string) {
	if s.ai != nil {
		s.ai.ClearChat(sessionID)
	}
}

func (s *Service) GenerateRule(ctx context.Context, req *types.RuleGenRequest) (*types.RuleGenResponse, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI service not available")
	}
	return s.ai.GenerateRule(ctx, req, s.policy.Engine(), s.telemetry.RawStore())
}

func (s *Service) ExplainEvent(ctx context.Context, req *types.ExplainRequest, eventID string) (*types.ExplainResponse, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI service not available")
	}
	record, ok := s.telemetry.Get(eventID)
	if !ok {
		return nil, fmt.Errorf("event %s not found", eventID)
	}
	return s.ai.ExplainEvent(ctx, req, record.Raw, s.policy.Engine(), s.telemetry.RawStore(), s.telemetry.Profiles(), s.telemetry.Workloads(), s.telemetry.ProcessTree(), s.stats)
}

func (s *Service) Analyze(ctx context.Context, req *types.AnalyzeRequest) (*types.AnalyzeResponse, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI service not available")
	}
	return s.ai.Analyze(ctx, req, s.telemetry.Profiles(), s.telemetry.Workloads(), s.policy.Engine(), s.stats, s.telemetry.RawStore(), s.telemetry.ProcessTree())
}

func (s *Service) AskAboutInsight(ctx context.Context, req *types.AskInsightRequest) (*types.AskInsightResponse, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("AI service not available")
	}
	return s.ai.AskAboutInsight(ctx, req)
}

func (s *Service) StartSentinel(cfg internalconfig.Config) {
	if !s.IsEnabled() || s.sentinel != nil {
		return
	}
	snt := sentinel.NewSentinel(s.ai, s.policy.Engine(), s.telemetry.RawStore(), s.telemetry.Profiles())
	schedule := sentinel.ScheduleConfig{}
	if d, err := time.ParseDuration(cfg.Sentinel.TestingPromotion); err == nil && d > 0 {
		schedule.TestingPromotion = d
	}
	if d, err := time.ParseDuration(cfg.Sentinel.Anomaly); err == nil && d > 0 {
		schedule.Anomaly = d
	}
	if d, err := time.ParseDuration(cfg.Sentinel.RuleOptimization); err == nil && d > 0 {
		schedule.RuleOptimization = d
	}
	if d, err := time.ParseDuration(cfg.Sentinel.DailyReport); err == nil && d > 0 {
		schedule.DailyReport = d
	}
	s.sentinel = snt.WithSchedule(schedule)
	s.sentinel.Start()
}

func (s *Service) StopSentinel() {
	if s.sentinel != nil {
		s.sentinel.Stop()
		s.sentinel = nil
	}
}

func (s *Service) Sentinel() *sentinel.Sentinel {
	return s.sentinel
}
