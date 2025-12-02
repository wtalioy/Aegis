package ui

import (
	"testing"
	"time"

	"eulerguard/pkg/types"
)

func TestStatsAlertDeduplication(t *testing.T) {
	stats := NewStats()
	stats.dedupWindow = 50 * time.Millisecond

	alert := types.Alert{
		RuleName:    "test",
		ProcessName: "bash",
		CgroupID:    "123",
		Action:      "alert",
	}

	stats.AddAlert(alert)
	stats.AddAlert(alert)

	if count := stats.AlertCount(); count != 1 {
		t.Fatalf("expected 1 alert, got %d", count)
	}
	if total := stats.TotalAlertCount(); total != 1 {
		t.Fatalf("expected total alert count 1, got %d", total)
	}

	stats.alertsMu.Lock()
	for key := range stats.alertDedup {
		stats.alertDedup[key] = time.Now().Add(-2 * stats.dedupWindow)
	}
	stats.alertsMu.Unlock()

	stats.AddAlert(alert)

	if count := stats.AlertCount(); count != 2 {
		t.Fatalf("expected 2 alerts after window, got %d", count)
	}
	if total := stats.TotalAlertCount(); total != 2 {
		t.Fatalf("expected total alert count 2, got %d", total)
	}
}
