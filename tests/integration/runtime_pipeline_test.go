package integration_test

import (
	"path/filepath"
	"testing"
	"time"

	"aegis/internal/app"
	internalconfig "aegis/internal/platform/config"
	"aegis/internal/policy"
	"aegis/internal/telemetry"
	"aegis/tests/helpers"
)

func TestRuntimeIngestPipelinePublishesEventsAndAlerts(t *testing.T) {
	cfg := internalconfig.Default(t.TempDir())
	cfg.Analysis.Mode = "disabled"
	cfg.Policy.RulesPath = filepath.Join(t.TempDir(), "rules.yaml")

	runtime := app.NewRuntime(cfg, filepath.Join(t.TempDir(), "config.yaml"))
	if err := runtime.Policy().Bootstrap([]policy.Rule{
		helpers.ActiveFileRule("watch-file", "/tmp/watch", policy.ActionAlert),
	}); err != nil {
		t.Fatalf("bootstrap rules: %v", err)
	}

	alertSub := runtime.AlertStream().Subscribe(10)
	defer alertSub.Cancel()
	eventSub := runtime.EventStream().Subscribe(10)
	defer eventSub.Cancel()

	event, decision, err := runtime.IngestPipeline().ProcessRawSample(
		helpers.RawFileSample(4200, 99, "bash", "/tmp/watch", 0, 1, 1, false),
	)
	if err != nil {
		t.Fatalf("process raw sample: %v", err)
	}
	if event == nil || event.ID == "" {
		t.Fatalf("expected ingested event with id, got %+v", event)
	}
	if got := decision.Type; got == "" || len(decision.Alerts) != 1 {
		t.Fatalf("expected alerting decision, got %+v", decision)
	}

	select {
	case published := <-eventSub.C:
		if published.ID != event.ID {
			t.Fatalf("expected event stream to publish %s, got %s", event.ID, published.ID)
		}
	default:
		t.Fatal("expected event publication")
	}

	select {
	case alert := <-alertSub.C:
		if alert.RuleName != "watch-file" {
			t.Fatalf("unexpected alert payload: %+v", alert)
		}
	default:
		t.Fatal("expected alert publication")
	}

	if got := len(runtime.Stats().Alerts()); got != 1 {
		t.Fatalf("expected stats alert buffer to contain alert, got %d", got)
	}
}

func TestSettingsServiceRejectsInvalidConfig(t *testing.T) {
	cfg := internalconfig.Default(t.TempDir())
	runtime := app.NewRuntime(cfg, filepath.Join(t.TempDir(), "config.yaml"))

	invalid := runtime.Settings().Get()
	invalid.Server.Port = 0

	if _, err := runtime.Settings().Update(invalid); err == nil {
		t.Fatal("expected invalid config update to fail")
	}
}

func TestSettingsServiceHotReloadsPolicyThresholds(t *testing.T) {
	cfg := internalconfig.Default(t.TempDir())
	cfg.Analysis.Mode = "disabled"
	cfg.Policy.RulesPath = filepath.Join(t.TempDir(), "rules.yaml")
	cfg.Policy.PromotionMinObservationMinutes = 60
	cfg.Policy.PromotionMinHits = 10

	runtime := app.NewRuntime(cfg, filepath.Join(t.TempDir(), "config.yaml"))
	if err := runtime.Policy().Bootstrap([]policy.Rule{
		{
			Name:        "watch-file",
			Description: "watch-file",
			Severity:    "warning",
			Action:      policy.ActionAlert,
			Type:        policy.RuleTypeFile,
			State:       policy.RuleStateTesting,
			Match: policy.MatchCondition{
				Filename: "/tmp/watch",
			},
		},
	}); err != nil {
		t.Fatalf("bootstrap rules: %v", err)
	}

	if _, _, err := runtime.IngestPipeline().ProcessRawSample(
		helpers.RawFileSample(4200, 99, "bash", "/tmp/watch", 0, 1, 1, false),
	); err != nil {
		t.Fatalf("process raw sample: %v", err)
	}

	readiness, _, err := runtime.Policy().Validation("watch-file")
	if err != nil {
		t.Fatalf("validation before reload: %v", err)
	}
	if readiness.IsReady {
		t.Fatalf("expected rule to remain not ready before hot reload, got %+v", readiness)
	}

	updated := runtime.Settings().Get()
	updated.Policy.PromotionMinObservationMinutes = 0
	updated.Policy.PromotionMinHits = 1

	result, err := runtime.Settings().Update(updated)
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if result.RestartRequired {
		t.Fatalf("expected policy threshold update to avoid restart, got %+v", result)
	}

	readiness, _, err = runtime.Policy().Validation("watch-file")
	if err != nil {
		t.Fatalf("validation after reload: %v", err)
	}
	if !readiness.IsReady {
		t.Fatalf("expected hot-reloaded thresholds to apply immediately, got %+v", readiness)
	}
}

func TestRuntimeIngestPipelineAllowsExecRuleWithoutAlert(t *testing.T) {
	cfg := internalconfig.Default(t.TempDir())
	cfg.Analysis.Mode = "disabled"
	cfg.Policy.RulesPath = filepath.Join(t.TempDir(), "rules.yaml")

	runtime := app.NewRuntime(cfg, filepath.Join(t.TempDir(), "config.yaml"))
	if err := runtime.Policy().Bootstrap([]policy.Rule{
		{
			Name:        "allow bash",
			Description: "allow bash",
			Severity:    "info",
			Action:      policy.ActionAllow,
			Type:        policy.RuleTypeExec,
			State:       policy.RuleStateProduction,
			Match: policy.MatchCondition{
				ProcessName:     "bash",
				ProcessNameType: policy.MatchTypeExact,
			},
		},
	}); err != nil {
		t.Fatalf("bootstrap rules: %v", err)
	}

	event, decision, err := runtime.IngestPipeline().ProcessRawSample(
		helpers.RawExecSample(5150, 5000, 88, "bash", "sshd", "/bin/bash", "bash -c id", false),
	)
	if err != nil {
		t.Fatalf("process raw sample: %v", err)
	}
	if event == nil || event.Type != telemetry.EventTypeExec {
		t.Fatalf("expected exec event, got %+v", event)
	}
	if decision.Type != policy.DecisionAllow {
		t.Fatalf("expected allow decision, got %+v", decision)
	}
	if len(decision.Alerts) != 0 || len(runtime.Stats().Alerts()) != 0 {
		t.Fatalf("expected allow path to avoid alerts, got decision=%+v stats=%+v", decision, runtime.Stats().Alerts())
	}
}

func TestRuntimeIngestPipelineCreatesSyntheticKernelBlockAlerts(t *testing.T) {
	cfg := internalconfig.Default(t.TempDir())
	cfg.Analysis.Mode = "disabled"
	cfg.Policy.RulesPath = filepath.Join(t.TempDir(), "rules.yaml")

	runtime := app.NewRuntime(cfg, filepath.Join(t.TempDir(), "config.yaml"))
	if err := runtime.Policy().Bootstrap(nil); err != nil {
		t.Fatalf("bootstrap rules: %v", err)
	}

	fileEvent, fileDecision, err := runtime.IngestPipeline().ProcessRawSample(
		helpers.RawFileSample(4200, 99, "bash", "/tmp/watch", 0, 1, 1, true),
	)
	if err != nil {
		t.Fatalf("process blocked file sample: %v", err)
	}
	if fileEvent == nil || fileDecision.Type != policy.DecisionBlock || len(fileDecision.Alerts) != 1 {
		t.Fatalf("unexpected blocked file decision: event=%+v decision=%+v", fileEvent, fileDecision)
	}
	if fileDecision.Alerts[0].RuleName != "Kernel Blocked File Access" || fileDecision.Alerts[0].Severity != "critical" {
		t.Fatalf("unexpected blocked file alert: %+v", fileDecision.Alerts[0])
	}

	connectEvent, connectDecision, err := runtime.IngestPipeline().ProcessRawSample(
		helpers.RawConnectSample(4200, 99, "bash", "192.168.1.10", 2, 4444, true),
	)
	if err != nil {
		t.Fatalf("process blocked connect sample: %v", err)
	}
	if connectEvent == nil || connectDecision.Type != policy.DecisionBlock || len(connectDecision.Alerts) != 1 {
		t.Fatalf("unexpected blocked connect decision: event=%+v decision=%+v", connectEvent, connectDecision)
	}
	if connectDecision.Alerts[0].RuleName != "Kernel Blocked Connection" || connectDecision.Alerts[0].Severity != "critical" {
		t.Fatalf("unexpected blocked connect alert: %+v", connectDecision.Alerts[0])
	}

	if got := len(runtime.Stats().Alerts()); got != 2 {
		t.Fatalf("expected synthetic alerts to be tracked in stats, got %d", got)
	}
}

func TestRuntimeTelemetryQueryRemainsStableAcrossPagesAfterMultipleIngests(t *testing.T) {
	cfg := internalconfig.Default(t.TempDir())
	cfg.Analysis.Mode = "disabled"
	cfg.Policy.RulesPath = filepath.Join(t.TempDir(), "rules.yaml")

	runtime := app.NewRuntime(cfg, filepath.Join(t.TempDir(), "config.yaml"))
	if err := runtime.Policy().Bootstrap(nil); err != nil {
		t.Fatalf("bootstrap rules: %v", err)
	}

	samples := [][]byte{
		helpers.RawExecSample(100, 1, 9, "bash", "init", "/bin/bash", "bash -c id", false),
		helpers.RawFileSample(101, 9, "cat", "/tmp/one", 0, 2, 1, false),
		helpers.RawConnectSample(102, 9, "curl", "10.0.0.5", 2, 443, false),
	}
	for _, sample := range samples {
		if _, _, err := runtime.IngestPipeline().ProcessRawSample(sample); err != nil {
			t.Fatalf("process raw sample: %v", err)
		}
		time.Sleep(time.Millisecond)
	}

	pageOne := runtime.Telemetry().Query(telemetry.Query{Page: 1, Limit: 2})
	pageTwo := runtime.Telemetry().Query(telemetry.Query{Page: 2, Limit: 2})

	if pageOne.Total != 3 || pageOne.TotalPages != 2 || len(pageOne.Events) != 2 {
		t.Fatalf("unexpected page one result: %+v", pageOne)
	}
	if len(pageTwo.Events) != 1 {
		t.Fatalf("unexpected page two result: %+v", pageTwo)
	}
	if pageOne.Events[0].Type != telemetry.EventTypeExec || pageOne.Events[1].Type != telemetry.EventTypeFile {
		t.Fatalf("expected page one ordering to remain stable, got %+v", pageOne.Events)
	}
	if pageTwo.Events[0].Type != telemetry.EventTypeConnect {
		t.Fatalf("expected page two to contain final event, got %+v", pageTwo.Events)
	}
}
