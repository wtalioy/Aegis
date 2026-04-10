package service

import (
	"context"

	"aegis/internal/analysis/types"
	"aegis/internal/platform/ai/analysis"
	"aegis/internal/platform/ai/snapshot"
	"aegis/internal/platform/storage"
	"aegis/internal/policy"
	"aegis/internal/shared/metrics"
	"aegis/internal/telemetry/proc"
	"aegis/internal/telemetry/workload"
)

// Analyze delegates to the analysis subpackage while keeping the Service API stable.
func (s *Service) Analyze(
	ctx context.Context,
	req *types.AnalyzeRequest,
	profileReg *proc.ProfileRegistry,
	workloadReg *workload.Registry,
	ruleEngine *policy.Engine,
	statsProvider metrics.StatsProvider,
	store storage.EventStore,
	processTree *proc.ProcessTree,
) (*types.AnalyzeResponse, error) {
	// Build snapshot state for context
	var snapshotState *snapshot.SystemState
	if statsProvider != nil && workloadReg != nil && store != nil {
		snap := snapshot.NewSnapshot(statsProvider, workloadReg, store, processTree)
		result := snap.BuildWithoutAncestors() // Don't need ancestors for analysis
		snapshotState = &result.State
	}
	return analysis.Analyze(ctx, s.provider, req, profileReg, workloadReg, ruleEngine, snapshotState)
}

// ExplainEvent delegates to the analysis subpackage.
func (s *Service) ExplainEvent(
	ctx context.Context,
	req *types.ExplainRequest,
	event *storage.Event,
	ruleEngine *policy.Engine,
	store storage.EventStore,
	profileReg *proc.ProfileRegistry,
	workloadReg *workload.Registry,
	processTree *proc.ProcessTree,
	statsProvider metrics.StatsProvider,
) (*types.ExplainResponse, error) {
	// Build snapshot state for context
	var snapshotState *snapshot.SystemState
	if statsProvider != nil && workloadReg != nil && store != nil {
		snap := snapshot.NewSnapshot(statsProvider, workloadReg, store, processTree)
		result := snap.BuildWithoutAncestors() // Ancestors built on-demand in BuildExplainContext
		snapshotState = &result.State
	}
	return analysis.ExplainEvent(ctx, s.provider, req, event, ruleEngine, store, profileReg, workloadReg, processTree, snapshotState)
}

// GenerateRule delegates to the analysis subpackage.
func (s *Service) GenerateRule(ctx context.Context, req *types.RuleGenRequest, ruleEngine *policy.Engine, store storage.EventStore) (*types.RuleGenResponse, error) {
	return analysis.GenerateRule(ctx, s.provider, req, ruleEngine, store)
}
