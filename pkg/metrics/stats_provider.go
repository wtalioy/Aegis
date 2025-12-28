package metrics

import "aegis/pkg/apimodel"

// StatsProvider is the minimal contract for exposing high-level event metrics
// and alerts. It is implemented by server.Stats and storage.Stats.
type StatsProvider interface {
	// Rates returns per-second execution/file/network event rates.
	Rates() (exec, file, net int64)

	// WorkloadCount returns the current number of active workloads.
	WorkloadCount() int

	// TotalAlertCount returns the total number of alerts ever recorded.
	TotalAlertCount() int64

	// Alerts returns the current list of alerts (copied).
	Alerts() []apimodel.Alert
}
