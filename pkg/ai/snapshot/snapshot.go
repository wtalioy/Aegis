package snapshot

import (
	"time"

	"aegis/pkg/metrics"
	"aegis/pkg/proc"
	"aegis/pkg/storage"
	"aegis/pkg/workload"
)

const (
	MaxAlertSummaries    = 10
	MaxActivitySummaries = 8
	RecentEventWindow    = 5 * time.Minute
	MaxEventsPerType     = 1000
	MaxAncestorDepth     = 7
)

// Snapshot encapsulates dependencies for building snapshots
type Snapshot struct {
	stats       metrics.StatsProvider
	workloadReg *workload.Registry
	store       storage.EventStore
	processTree *proc.ProcessTree
}

func NewSnapshot(statsProvider metrics.StatsProvider, workloadReg *workload.Registry, store storage.EventStore, processTree *proc.ProcessTree) *Snapshot {
	return &Snapshot{
		stats:       statsProvider,
		workloadReg: workloadReg,
		store:       store,
		processTree: processTree,
	}
}

type Result struct {
	State              SystemState
	ProcessKeyToChain  map[string]string // comm|parentComm -> ancestor chain
	ProcessNameToChain map[string]string // process name -> ancestor chain
}

func (s *Snapshot) Build() Result {
	execRate, fileRate, netRate := s.stats.Rates()
	state := s.buildBaseState(execRate, fileRate, netRate)

	state.TopWorkloads = s.buildTopWorkloads()

	alerts := s.stats.Alerts()
	state.RecentAlerts = s.deduplicateAlerts(alerts)

	execEvents, recentActivity := s.buildRecentActivity()
	state.RecentProcesses = recentActivity.Processes
	state.RecentConnections = recentActivity.Connections
	state.RecentFileAccess = recentActivity.Files

	processKeyToChain, processNameToChain := s.buildAncestorChainMaps(execEvents, alerts)

	return Result{
		State:              state,
		ProcessKeyToChain:  processKeyToChain,
		ProcessNameToChain: processNameToChain,
	}
}

func (s *Snapshot) BuildWithoutAncestors() Result {
	execRate, fileRate, netRate := s.stats.Rates()
	state := s.buildBaseState(execRate, fileRate, netRate)

	state.TopWorkloads = s.buildTopWorkloads()

	alerts := s.stats.Alerts()
	state.RecentAlerts = s.deduplicateAlerts(alerts)

	_, recentActivity := s.buildRecentActivity()
	state.RecentProcesses = recentActivity.Processes
	state.RecentConnections = recentActivity.Connections
	state.RecentFileAccess = recentActivity.Files

	// Return empty maps for ancestor chains - they'll be built on-demand when needed
	return Result{
		State:              state,
		ProcessKeyToChain:  make(map[string]string),
		ProcessNameToChain: make(map[string]string),
	}
}
