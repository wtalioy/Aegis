package telemetry_test

import (
	"testing"
	"time"

	"aegis/internal/platform/events"
	"aegis/internal/telemetry"
	"aegis/internal/telemetry/proc"
	"aegis/internal/telemetry/workload"
	"aegis/tests/helpers"
)

func TestTelemetryIngestAndQueryUpdatesProjections(t *testing.T) {
	processTree := proc.NewProcessTree(time.Minute, 1000, 16)
	workloads := workload.NewRegistry(100)
	profiles := proc.NewProfileRegistry()
	profiles.GetOrCreateProfile(4200, time.Now(), "bash -c whoami", nil)

	service := telemetry.NewService(100, 100, processTree, workloads, profiles)

	execRecord, err := events.DecodeSample(helpers.RawExecSample(4200, 4100, 77, "bash", "sshd", "/bin/bash", "bash -c whoami", false))
	if err != nil {
		t.Fatalf("decode exec: %v", err)
	}
	execResult, err := service.Ingest(execRecord)
	if err != nil {
		t.Fatalf("ingest exec: %v", err)
	}

	fileRecord, err := events.DecodeSample(helpers.RawFileSample(4200, 77, "bash", "/etc/hosts", 0, 1, 1, false))
	if err != nil {
		t.Fatalf("decode file: %v", err)
	}
	fileResult, err := service.Ingest(fileRecord)
	if err != nil {
		t.Fatalf("ingest file: %v", err)
	}

	connectRecord, err := events.DecodeSample(helpers.RawConnectSample(4200, 77, "bash", "10.0.0.8", 2, 443, false))
	if err != nil {
		t.Fatalf("decode connect: %v", err)
	}
	connectResult, err := service.Ingest(connectRecord)
	if err != nil {
		t.Fatalf("ingest connect: %v", err)
	}

	execEvent := execResult.Event
	fileEvent := fileResult.Event
	connectEvent := connectResult.Event
	if execEvent.ID == "" || fileEvent.ID == "" || connectEvent.ID == "" {
		t.Fatal("expected stable event ids for all ingested events")
	}

	if info, ok := processTree.GetProcess(4200); !ok || info.Comm != "bash" {
		t.Fatalf("expected process tree to contain ingested exec, got %+v", info)
	}

	workloadMeta := workloads.Get(77)
	if workloadMeta == nil {
		t.Fatal("expected workload metadata to be created")
	}
	if workloadMeta.ExecCount != 1 || workloadMeta.FileCount != 1 || workloadMeta.ConnectCount != 1 {
		t.Fatalf("unexpected workload counters: %+v", workloadMeta)
	}

	profile, ok := profiles.GetProfile(4200)
	if !ok {
		t.Fatal("expected process profile to exist")
	}
	if profile.Dynamic.ExecCount != 1 || profile.Dynamic.FileOpenCount != 1 || profile.Dynamic.NetConnectCount != 1 {
		t.Fatalf("unexpected profile counters: %+v", profile.Dynamic)
	}

	result := service.Query(telemetry.Query{
		Filter: telemetry.Filter{
			Types:     []telemetry.EventType{telemetry.EventTypeExec, telemetry.EventTypeConnect},
			Processes: []string{"bash"},
		},
		Page:  1,
		Limit: 10,
	})
	if result.Total != 2 {
		t.Fatalf("expected 2 filtered events, got %d", result.Total)
	}
	if result.TypeCounts.Exec != 1 || result.TypeCounts.Connect != 1 || result.TypeCounts.File != 0 {
		t.Fatalf("unexpected type counts: %+v", result.TypeCounts)
	}

	latest := service.Latest(2)
	if len(latest) != 2 {
		t.Fatalf("expected latest query to return 2 events, got %d", len(latest))
	}
	if latest[0].Type != telemetry.EventTypeConnect || latest[1].Type != telemetry.EventTypeFile {
		t.Fatalf("unexpected latest event order: %+v", latest)
	}
}
