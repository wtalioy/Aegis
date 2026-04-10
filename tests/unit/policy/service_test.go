package policy_test

import (
	"testing"
	"time"

	"aegis/internal/platform/events"
	"aegis/internal/platform/storage"
	"aegis/internal/policy"
	"aegis/internal/telemetry"
	"aegis/tests/fakes"
	"aegis/tests/helpers"
)

func TestRuleServiceLifecyclePersistsAndPromotesRules(t *testing.T) {
	repo := fakes.NewRuleRepository([]policy.Rule{
		helpers.ActiveFileRule("existing", "/tmp/one", policy.ActionAlert),
	})
	syncer := &fakes.KernelSync{}
	service := policy.NewService(repo, syncer, 60, 10)

	if err := service.Load(); err != nil {
		t.Fatalf("load rules: %v", err)
	}
	if syncer.SyncCall != 1 {
		t.Fatalf("expected kernel sync on initial load, got %d", syncer.SyncCall)
	}

	created, err := service.Create(policy.Rule{
		Name:        "draft-rule",
		Description: "created during test",
		Severity:    "warning",
		Action:      policy.ActionAlert,
		Type:        policy.RuleTypeFile,
		Match: policy.MatchCondition{
			Filename: "/tmp/two",
		},
	})
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}
	if created.State != policy.RuleStateTesting {
		t.Fatalf("expected created rule to default to testing, got %s", created.State)
	}

	if err := service.Promote("draft-rule"); err != nil {
		t.Fatalf("promote rule: %v", err)
	}
	promoted, ok := service.Get("draft-rule")
	if !ok || !promoted.IsProduction() {
		t.Fatalf("expected promoted rule to be production, got %+v", promoted)
	}

	if err := service.Delete("existing"); err != nil {
		t.Fatalf("delete rule: %v", err)
	}
	if _, ok := service.Get("existing"); ok {
		t.Fatal("expected deleted rule to be removed")
	}
}

func TestEvaluateTestingRuleRecordsHitWithoutAlert(t *testing.T) {
	repo := fakes.NewRuleRepository([]policy.Rule{
		{
			Name:        "testing file",
			Description: "testing file",
			Severity:    "info",
			Action:      policy.ActionAlert,
			Type:        policy.RuleTypeFile,
			State:       policy.RuleStateTesting,
			Match: policy.MatchCondition{
				Filename: "/tmp/watch",
			},
		},
	})
	service := policy.NewService(repo, &fakes.KernelSync{}, 60, 10)
	if err := service.Load(); err != nil {
		t.Fatalf("load rules: %v", err)
	}

	sample := helpers.RawFileSample(99, 7, "cat", "/tmp/watch", 0, 1, 1, false)
	fileEvent, err := events.DecodeFileOpenEvent(sample)
	if err != nil {
		t.Fatalf("decode file sample: %v", err)
	}
	raw := &storage.Event{Timestamp: time.Now(), Data: fileEvent}
	decision := service.Evaluate(&telemetry.Record{
		Event: telemetry.Event{
			ID:          "event-1",
			Type:        telemetry.EventTypeFile,
			Timestamp:   time.Now(),
			PID:         99,
			CgroupID:    7,
			ProcessName: "cat",
			Filename:    "/tmp/watch",
		},
		Raw: raw,
	})

	if decision.Type != policy.DecisionTestingHit {
		t.Fatalf("expected testing hit decision, got %s", decision.Type)
	}
	if len(decision.Alerts) != 0 {
		t.Fatalf("expected no alerts for testing rule, got %+v", decision.Alerts)
	}
}
