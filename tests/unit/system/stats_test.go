package system_test

import (
	"testing"
	"time"

	"aegis/internal/system"
)

func TestStatsDeduplicatesAlertsWithinWindow(t *testing.T) {
	stats := system.NewStats(10, time.Minute)

	alert := system.Alert{
		RuleName:    "rule-1",
		ProcessName: "bash",
		CgroupID:    "123",
		Action:      "alert",
	}

	stats.AddAlert(alert)
	stats.AddAlert(alert)

	if got := len(stats.Alerts()); got != 1 {
		t.Fatalf("expected 1 alert after deduplication, got %d", got)
	}
	if got := stats.TotalAlertCount(); got != 1 {
		t.Fatalf("expected total alert count to remain deduplicated, got %d", got)
	}
}
