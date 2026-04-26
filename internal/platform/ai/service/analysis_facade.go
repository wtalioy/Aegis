package service

import (
	"context"

	"aegis/internal/analysis/types"
	"aegis/internal/platform/ai/analysis"
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
	provider, err := s.requireProvider()
	if err != nil {
		return nil, err
	}
	return analysis.Analyze(ctx, provider, req, profileReg, workloadReg, ruleEngine, s.snapshotState(runtimeDeps{
		statsProvider: statsProvider,
		workloadReg:   workloadReg,
		store:         store,
		processTree:   processTree,
	}))
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
	provider, err := s.requireProvider()
	if err != nil {
		return nil, err
	}
	return analysis.ExplainEvent(ctx, provider, req, event, ruleEngine, store, profileReg, workloadReg, processTree, s.snapshotState(runtimeDeps{
		statsProvider: statsProvider,
		workloadReg:   workloadReg,
		store:         store,
		processTree:   processTree,
	}))
}

// GenerateRule delegates to the analysis subpackage.
func (s *Service) GenerateRule(ctx context.Context, req *types.RuleGenRequest, ruleEngine *policy.Engine, store storage.EventStore) (*types.RuleGenResponse, error) {
	provider, err := s.requireProvider()
	if err != nil {
		return nil, err
	}
	return analysis.GenerateRule(ctx, provider, req, ruleEngine, store)
}
