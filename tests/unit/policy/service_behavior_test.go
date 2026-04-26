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

func fileRecord(t *testing.T, sample []byte, filename string, blocked bool) *telemetry.Record {
	t.Helper()

	fileEvent, err := events.DecodeFileOpenEvent(sample)
	if err != nil {
		t.Fatalf("decode file sample: %v", err)
	}

	return &telemetry.Record{
		Event: telemetry.Event{
			ID:          "file-event",
			Type:        telemetry.EventTypeFile,
			Timestamp:   time.Now(),
			PID:         fileEvent.Hdr.PID,
			CgroupID:    fileEvent.Hdr.CgroupID,
			ProcessName: "cat",
			Filename:    filename,
			Blocked:     blocked,
			Ino:         fileEvent.Ino,
			Dev:         fileEvent.Dev,
			Flags:       fileEvent.Flags,
		},
		Raw: &storage.Event{Timestamp: time.Now(), Data: fileEvent},
	}
}

func execRecord(t *testing.T, sample []byte) *telemetry.Record {
	t.Helper()

	record, err := events.DecodeSample(sample)
	if err != nil {
		t.Fatalf("decode exec sample: %v", err)
	}
	ingested, err := telemetry.NewService(10, 10, nil, nil, nil).Ingest(record)
	if err != nil {
		t.Fatalf("ingest exec sample: %v", err)
	}
	return ingested
}

func connectRecord(t *testing.T, sample []byte) *telemetry.Record {
	t.Helper()

	record, err := events.DecodeSample(sample)
	if err != nil {
		t.Fatalf("decode connect sample: %v", err)
	}
	ingested, err := telemetry.NewService(10, 10, nil, nil, nil).Ingest(record)
	if err != nil {
		t.Fatalf("ingest connect sample: %v", err)
	}
	return ingested
}

func TestPolicyService_CreatePromoteDeleteLifecycle(t *testing.T) {
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

func TestPolicyService_EvaluateTestingRuleRecordsHitWithoutAlert(t *testing.T) {
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

	decision := service.Evaluate(fileRecord(t, helpers.RawFileSample(99, 7, "cat", "/tmp/watch", 0, 1, 1, false), "/tmp/watch", false))

	if decision.Type != policy.DecisionTestingHit {
		t.Fatalf("expected testing hit decision, got %s", decision.Type)
	}
	if len(decision.Alerts) != 0 {
		t.Fatalf("expected no alerts for testing rule, got %+v", decision.Alerts)
	}
}

func TestPolicyService_EvaluateReturnsNoMatchForNilOrMissingRawPayload(t *testing.T) {
	service := policy.NewService(fakes.NewRuleRepository(nil), &fakes.KernelSync{}, 60, 10)
	if err := service.Bootstrap(nil); err != nil {
		t.Fatalf("bootstrap rules: %v", err)
	}

	if decision := service.Evaluate(nil); decision.Type != policy.DecisionNoMatch {
		t.Fatalf("expected nil record to yield no_match, got %s", decision.Type)
	}

	decision := service.Evaluate(&telemetry.Record{
		Event: telemetry.Event{
			Type:        telemetry.EventTypeExec,
			Timestamp:   time.Now(),
			PID:         44,
			CgroupID:    9,
			ProcessName: "bash",
		},
	})
	if decision.Type != policy.DecisionNoMatch {
		t.Fatalf("expected missing raw payload to yield no_match, got %s", decision.Type)
	}
}

func TestPolicyService_EvaluateExecAllowRuleReturnsAllowWithoutAlert(t *testing.T) {
	repo := fakes.NewRuleRepository([]policy.Rule{
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
	})
	service := policy.NewService(repo, &fakes.KernelSync{}, 60, 10)
	if err := service.Load(); err != nil {
		t.Fatalf("load rules: %v", err)
	}

	decision := service.Evaluate(execRecord(t, helpers.RawExecSample(5150, 5000, 88, "bash", "sshd", "/bin/bash", "bash -c id", false)))
	if decision.Type != policy.DecisionAllow {
		t.Fatalf("expected allow decision, got %s", decision.Type)
	}
	if decision.Rule == nil || decision.Rule.Name != "allow bash" {
		t.Fatalf("expected allow rule to be returned, got %+v", decision.Rule)
	}
	if len(decision.Alerts) != 0 {
		t.Fatalf("expected no alerts for allow rule, got %+v", decision.Alerts)
	}
}

func TestPolicyService_EvaluateExecBlockedWithoutRuleReturnsSyntheticCriticalAlert(t *testing.T) {
	service := policy.NewService(fakes.NewRuleRepository(nil), &fakes.KernelSync{}, 60, 10)
	if err := service.Bootstrap(nil); err != nil {
		t.Fatalf("bootstrap rules: %v", err)
	}

	decision := service.Evaluate(execRecord(t, helpers.RawExecSample(5150, 5000, 88, "bash", "sshd", "/bin/bash", "bash -c id", true)))
	if decision.Type != policy.DecisionBlock {
		t.Fatalf("expected block decision, got %s", decision.Type)
	}
	if len(decision.Alerts) != 1 {
		t.Fatalf("expected one synthetic alert, got %+v", decision.Alerts)
	}
	if got := decision.Alerts[0]; got.RuleName != "Kernel Blocked Execution" || got.Severity != "critical" || !got.Blocked {
		t.Fatalf("unexpected synthetic exec alert: %+v", got)
	}
}

func TestPolicyService_EvaluateFileAndConnectRulesCoverAlertAndBlockBranches(t *testing.T) {
	repo := fakes.NewRuleRepository([]policy.Rule{
		{
			Name:        "watch file",
			Description: "watch file",
			Severity:    "warning",
			Action:      policy.ActionAlert,
			Type:        policy.RuleTypeFile,
			State:       policy.RuleStateProduction,
			Match: policy.MatchCondition{
				Filename: "/tmp/watch",
			},
		},
		{
			Name:        "block 4444",
			Description: "block 4444",
			Severity:    "critical",
			Action:      policy.ActionBlock,
			Type:        policy.RuleTypeConnect,
			State:       policy.RuleStateProduction,
			Match: policy.MatchCondition{
				DestPort: 4444,
			},
		},
	})
	service := policy.NewService(repo, &fakes.KernelSync{}, 60, 10)
	if err := service.Load(); err != nil {
		t.Fatalf("load rules: %v", err)
	}

	fileDecision := service.Evaluate(fileRecord(t, helpers.RawFileSample(99, 7, "cat", "/tmp/watch", 0, 1, 1, false), "/tmp/watch", false))
	if fileDecision.Type != policy.DecisionAlert {
		t.Fatalf("expected file alert decision, got %s", fileDecision.Type)
	}
	if len(fileDecision.Alerts) != 1 || fileDecision.Alerts[0].RuleName != "watch file" {
		t.Fatalf("unexpected file alert payload: %+v", fileDecision.Alerts)
	}

	connectDecision := service.Evaluate(connectRecord(t, helpers.RawConnectSample(88, 7, "curl", "192.168.1.10", 2, 4444, false)))
	if connectDecision.Type != policy.DecisionBlock {
		t.Fatalf("expected connect block decision, got %s", connectDecision.Type)
	}
	if len(connectDecision.Alerts) != 1 || connectDecision.Alerts[0].Action != string(policy.ActionBlock) {
		t.Fatalf("unexpected connect alert payload: %+v", connectDecision.Alerts)
	}
}

func TestPolicyService_TestingRulesAndThresholdUpdatesReuseExistingHits(t *testing.T) {
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

	oldHit := fileRecord(t, helpers.RawFileSample(99, 7, "cat", "/tmp/watch", 0, 1, 1, false), "/tmp/watch", false)
	oldHit.Event.Timestamp = time.Now().Add(-2 * time.Hour)
	oldHit.Raw.Timestamp = oldHit.Event.Timestamp
	if decision := service.Evaluate(oldHit); decision.Type != policy.DecisionTestingHit {
		t.Fatalf("expected testing hit, got %s", decision.Type)
	}

	newHit := fileRecord(t, helpers.RawFileSample(99, 7, "cat", "/tmp/watch", 0, 1, 1, false), "/tmp/watch", false)
	if decision := service.Evaluate(newHit); decision.Type != policy.DecisionTestingHit {
		t.Fatalf("expected testing hit, got %s", decision.Type)
	}

	items := service.TestingRules()
	if len(items) != 1 || items[0].Validation.IsReady {
		t.Fatalf("expected rule to remain not ready before threshold update, got %+v", items)
	}
	if items[0].Stats.Hits != 2 {
		t.Fatalf("expected two recorded hits, got %+v", items[0].Stats)
	}

	service.UpdateThresholds(0, 2)

	items = service.TestingRules()
	if len(items) != 1 || !items[0].Validation.IsReady {
		t.Fatalf("expected threshold update to mark rule ready, got %+v", items)
	}
}

func TestPolicyService_UpdatePreservesLifecycleFieldsWhenOmitted(t *testing.T) {
	createdAt := time.Now().Add(-4 * time.Hour).UTC().Truncate(time.Second)
	deployedAt := createdAt.Add(time.Hour)
	promotedAt := createdAt.Add(2 * time.Hour)
	repo := fakes.NewRuleRepository([]policy.Rule{
		{
			Name:        "watch file",
			Description: "watch file",
			Severity:    "warning",
			Action:      policy.ActionAlert,
			Type:        policy.RuleTypeFile,
			State:       policy.RuleStateProduction,
			CreatedAt:   createdAt,
			DeployedAt:  &deployedAt,
			PromotedAt:  &promotedAt,
			Match: policy.MatchCondition{
				Filename: "/tmp/watch",
			},
		},
	})
	service := policy.NewService(repo, &fakes.KernelSync{}, 60, 10)
	if err := service.Load(); err != nil {
		t.Fatalf("load rules: %v", err)
	}

	updated, err := service.Update("watch file", policy.Rule{
		Description: "updated description",
		Severity:    "critical",
		Action:      policy.ActionAlert,
		Type:        policy.RuleTypeFile,
		Match: policy.MatchCondition{
			Filename: "/tmp/updated",
		},
	})
	if err != nil {
		t.Fatalf("update rule: %v", err)
	}

	if !updated.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected createdAt to be preserved, got %v", updated.CreatedAt)
	}
	if updated.DeployedAt == nil || !updated.DeployedAt.Equal(deployedAt) {
		t.Fatalf("expected deployedAt to be preserved, got %+v", updated.DeployedAt)
	}
	if updated.PromotedAt == nil || !updated.PromotedAt.Equal(promotedAt) {
		t.Fatalf("expected promotedAt to be preserved, got %+v", updated.PromotedAt)
	}
	if updated.State != policy.RuleStateProduction {
		t.Fatalf("expected rule state to be preserved, got %s", updated.State)
	}
}
