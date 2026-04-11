package httpapi

import (
	"context"

	"aegis/internal/analysis"
	analysissentinel "aegis/internal/analysis/sentinel"
	analysistypes "aegis/internal/analysis/types"
	"aegis/internal/policy"
	"aegis/internal/shared/stream"
	"aegis/internal/system"
	"aegis/internal/telemetry"
	"aegis/internal/telemetry/proc"
)

type EventStream interface {
	Subscribe(buffer int) stream.Subscription[telemetry.Event]
}

type AlertStream interface {
	Subscribe(buffer int) stream.Subscription[system.Alert]
}

type TelemetryService interface {
	Query(telemetry.Query) telemetry.PageResult
	Get(id string) (*telemetry.Record, bool)
	ProcessTree() *proc.ProcessTree
}

type PolicyService interface {
	TestingRules() []policy.TestingRuleStatus
	Validation(name string) (policy.PromotionReadiness, policy.TestingStats, error)
	List() []policy.Rule
	Create(rule policy.Rule) (policy.Rule, error)
	Get(name string) (*policy.Rule, bool)
	Update(name string, rule policy.Rule) (policy.Rule, error)
	Delete(name string) error
	Promote(name string) error
}

type AnalysisService interface {
	Status() analysistypes.StatusDTO
	Diagnose(ctx context.Context) (*analysistypes.DiagnosisResult, error)
	Chat(ctx context.Context, sessionID, message string) (*analysistypes.ChatResponse, error)
	ChatStream(ctx context.Context, sessionID, message string) (<-chan analysistypes.ChatStreamToken, error)
	ChatHistory(sessionID string) []analysistypes.Message
	ClearChat(sessionID string)
	GenerateRule(ctx context.Context, req *analysistypes.RuleGenRequest) (*analysistypes.RuleGenResponse, error)
	ExplainEvent(ctx context.Context, req *analysistypes.ExplainRequest, eventID string) (*analysistypes.ExplainResponse, error)
	Analyze(ctx context.Context, req *analysistypes.AnalyzeRequest) (*analysistypes.AnalyzeResponse, error)
	AskAboutInsight(ctx context.Context, req *analysistypes.AskInsightRequest) (*analysistypes.AskInsightResponse, error)
	Sentinel() *analysissentinel.Sentinel
}

type SettingsService interface {
	Get() system.Settings
	Update(cfg system.Settings) (system.UpdateResult, error)
}

type StatsService interface {
	Rates() (exec, file, net int64)
	WorkloadCount() int
	TotalAlertCount() int64
	Alerts() []system.Alert
}

type ProbeStatusService interface {
	ProbeStatus() system.ProbeStatus
}

type Dependencies struct {
	Telemetry   TelemetryService
	Policy      PolicyService
	Analysis    AnalysisService
	Settings    SettingsService
	Stats       StatsService
	ProbeStatus ProbeStatusService
	EventStream EventStream
	AlertStream AlertStream
}

type runtimeView interface {
	Settings() *system.SettingsService
	Stats() *system.Stats
	Telemetry() *telemetry.Service
	Policy() *policy.Service
	Analysis() *analysis.Service
	EventStream() *stream.Hub[telemetry.Event]
	AlertStream() *stream.Hub[system.Alert]
	ProbeStatus() system.ProbeStatus
}

func DependenciesFromRuntime(runtime runtimeView) Dependencies {
	deps := Dependencies{
		Telemetry:   runtime.Telemetry(),
		Policy:      runtime.Policy(),
		Settings:    runtime.Settings(),
		Stats:       runtime.Stats(),
		ProbeStatus: runtime,
		EventStream: runtime.EventStream(),
		AlertStream: runtime.AlertStream(),
	}
	if analysisService := runtime.Analysis(); analysisService != nil {
		deps.Analysis = analysisService
	}
	return deps
}
