package integration_test

import (
	"path/filepath"
	"testing"

	"aegis/internal/app"
	internalconfig "aegis/internal/platform/config"
	"aegis/internal/policy"
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
