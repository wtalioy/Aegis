package integration_test

import (
	"testing"
	"time"

	"aegis/internal/platform/events"
	"aegis/internal/policy"
	"aegis/internal/system"
	"aegis/internal/telemetry"
	"aegis/internal/telemetry/proc"
	"aegis/internal/telemetry/workload"
	"aegis/tests/fakes"
	"aegis/tests/helpers"
)

func TestTelemetryPipeline_ProcessesExecFileAndConnectEventsEndToEnd(t *testing.T) {
	processTree := proc.NewProcessTree(time.Minute, 1000, 16)
	workloads := workload.NewRegistry(100)
	profiles := proc.NewProfileRegistry()
	profiles.GetOrCreateProfile(5150, time.Now(), "bash -c id", nil)

	telemetryService := telemetry.NewService(100, 100, processTree, workloads, profiles)
	stats := system.NewStats(10, time.Second)
	stats.SetWorkloadCountFunc(workloads.Count)

	repo := fakes.NewRuleRepository([]policy.Rule{
		{
			Name:        "exec alert",
			Description: "alert on bash",
			Severity:    "warning",
			Action:      policy.ActionAlert,
			Type:        policy.RuleTypeExec,
			State:       policy.RuleStateProduction,
			Match: policy.MatchCondition{
				ProcessName:     "bash",
				ProcessNameType: policy.MatchTypeExact,
			},
		},
		{
			Name:        "file testing",
			Description: "test file access",
			Severity:    "info",
			Action:      policy.ActionAlert,
			Type:        policy.RuleTypeFile,
			State:       policy.RuleStateTesting,
			Match: policy.MatchCondition{
				Filename: "/tmp/watch",
			},
		},
		{
			Name:        "connect block",
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
	policyService := policy.NewService(repo, &fakes.KernelSync{}, 60, 10)
	if err := policyService.Load(); err != nil {
		t.Fatalf("load policy service: %v", err)
	}

	samples := [][]byte{
		helpers.RawExecSample(5150, 5000, 88, "bash", "sshd", "/bin/bash", "bash -c id", false),
		helpers.RawFileSample(5150, 88, "bash", "/tmp/watch", 0, 1, 1, false),
		helpers.RawConnectSample(5150, 88, "bash", "192.168.1.10", 2, 4444, true),
	}

	for _, sample := range samples {
		record, err := events.DecodeSample(sample)
		if err != nil {
			t.Fatalf("decode sample: %v", err)
		}
		ingested, err := telemetryService.Ingest(record)
		if err != nil {
			t.Fatalf("ingest sample: %v", err)
		}
		event := &ingested.Event
		switch event.Type {
		case telemetry.EventTypeExec:
			stats.RecordExec()
		case telemetry.EventTypeFile:
			stats.RecordFile()
		case telemetry.EventTypeConnect:
			stats.RecordConnect()
		}
		decision := policyService.Evaluate(ingested)
		for _, alert := range decision.Alerts {
			stats.AddAlert(alert)
			telemetryService.RecordAlert(event.CgroupID, alert.Blocked)
		}
	}

	execCount, fileCount, connectCount := stats.Counts()
	if execCount != 1 || fileCount != 1 || connectCount != 1 {
		t.Fatalf("unexpected event counts: exec=%d file=%d connect=%d", execCount, fileCount, connectCount)
	}

	if got := len(stats.Alerts()); got != 2 {
		t.Fatalf("expected 2 emitted alerts, got %d", got)
	}

	workloadMeta := workloads.Get(88)
	if workloadMeta == nil {
		t.Fatal("expected workload metadata for cgroup 88")
	}
	if workloadMeta.AlertCount != 2 || workloadMeta.BlockedCount != 1 {
		t.Fatalf("unexpected workload alert counters: %+v", workloadMeta)
	}

	profile, ok := profiles.GetProfile(5150)
	if !ok {
		t.Fatal("expected process profile to exist")
	}
	if profile.Dynamic.ExecCount != 1 || profile.Dynamic.FileOpenCount != 1 || profile.Dynamic.NetConnectCount != 1 {
		t.Fatalf("unexpected profile counters after pipeline: %+v", profile.Dynamic)
	}

	queryResult := telemetryService.Query(telemetry.Query{Page: 1, Limit: 10})
	if queryResult.Total != 3 {
		t.Fatalf("expected 3 events in telemetry store, got %d", queryResult.Total)
	}

	readiness, testingStats, err := policyService.Validation("file testing")
	if err != nil {
		t.Fatalf("testing rule validation: %v", err)
	}
	if testingStats.Hits != 1 {
		t.Fatalf("expected testing buffer hit to be recorded, got %+v", testingStats)
	}
	if readiness.IsReady {
		t.Fatalf("expected single-hit testing rule to remain not ready, got %+v", readiness)
	}
}
