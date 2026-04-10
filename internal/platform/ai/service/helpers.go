package service

import (
	"fmt"
	"time"

	"aegis/internal/analysis/types"
	"aegis/internal/platform/ai/providers"
	"aegis/internal/platform/ai/snapshot"
	"aegis/internal/platform/storage"
	"aegis/internal/shared/metrics"
	"aegis/internal/telemetry/proc"
	"aegis/internal/telemetry/workload"
)

func (s *Service) requireProvider() (providers.Provider, error) {
	if s == nil || s.provider == nil {
		return nil, fmt.Errorf("AI service is not available")
	}
	return s.provider, nil
}

func (s *Service) buildSnapshot(
	statsProvider metrics.StatsProvider,
	workloadReg *workload.Registry,
	store storage.EventStore,
	processTree *proc.ProcessTree,
) snapshot.Result {
	return snapshot.NewSnapshot(statsProvider, workloadReg, store, processTree).Build()
}

func (s *Service) appendUserAndAssistant(sessionID, userMessage, assistantMessage string, ts int64) {
	if s == nil || s.conversations == nil {
		return
	}
	s.conversations.AddMessage(sessionID, types.Message{Role: "user", Content: userMessage, Timestamp: ts})
	s.conversations.AddMessage(sessionID, types.Message{Role: "assistant", Content: assistantMessage, Timestamp: ts})
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}
