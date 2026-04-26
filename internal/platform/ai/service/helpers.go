package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"aegis/internal/analysis/types"
	"aegis/internal/platform/ai/providers"
	"aegis/internal/platform/ai/snapshot"
	"aegis/internal/platform/config"
	"aegis/internal/platform/storage"
	"aegis/internal/shared/metrics"
	"aegis/internal/telemetry/proc"
	"aegis/internal/telemetry/workload"
)

type runtimeDeps struct {
	statsProvider metrics.StatsProvider
	workloadReg   *workload.Registry
	store         storage.EventStore
	processTree   *proc.ProcessTree
}

func (s *Service) requireProvider() (providers.Provider, error) {
	if s == nil || s.provider == nil {
		return nil, fmt.Errorf("AI service is not available")
	}
	return s.provider, nil
}

func newProvider(opts config.AIOptions) (providers.Provider, error) {
	switch opts.Mode {
	case "ollama":
		return providers.NewOllamaProvider(opts.Ollama), nil
	case "openai":
		return providers.NewOpenAIProvider(opts.OpenAI), nil
	case "gemini":
		return providers.NewGeminiProvider(opts.Gemini), nil
	default:
		return nil, fmt.Errorf("unknown AI mode: %s", opts.Mode)
	}
}

func logProviderHealth(provider providers.Provider) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := provider.CheckHealth(ctx); err != nil {
		log.Printf("[AI] Warning: Provider health check failed: %v", err)
		return
	}
	log.Printf("[AI] Provider %s initialized successfully", provider.Name())
}

func (s *Service) providerAndSnapshot(deps runtimeDeps) (providers.Provider, snapshot.Result, error) {
	provider, err := s.requireProvider()
	if err != nil {
		return nil, snapshot.Result{}, err
	}
	return provider, snapshot.NewSnapshot(deps.statsProvider, deps.workloadReg, deps.store, deps.processTree).Build(), nil
}

func (s *Service) snapshotState(deps runtimeDeps) *snapshot.SystemState {
	if deps.statsProvider == nil || deps.workloadReg == nil || deps.store == nil {
		return nil
	}
	result := snapshot.NewSnapshot(deps.statsProvider, deps.workloadReg, deps.store, deps.processTree).BuildWithoutAncestors()
	return &result.State
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
