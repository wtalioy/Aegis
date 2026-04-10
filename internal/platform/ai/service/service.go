package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"aegis/internal/analysis/types"
	"aegis/internal/platform/ai/chat"
	"aegis/internal/platform/ai/diagnostics"
	"aegis/internal/platform/ai/providers"
	"aegis/internal/platform/config"
	"aegis/internal/platform/storage"
	"aegis/internal/shared/metrics"
	"aegis/internal/telemetry/proc"
	"aegis/internal/telemetry/workload"
)

type Service struct {
	provider      providers.Provider
	conversations *chat.Store
}

func NewClient(p providers.Provider) *Service {
	return &Service{
		provider:      p,
		conversations: chat.NewStore(),
	}
}

func NewService(opts config.AIOptions) (*Service, error) {
	var provider providers.Provider

	switch opts.Mode {
	case "ollama":
		provider = providers.NewOllamaProvider(opts.Ollama)
	case "openai":
		provider = providers.NewOpenAIProvider(opts.OpenAI)
	default:
		return nil, fmt.Errorf("unknown AI mode: %s", opts.Mode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := provider.CheckHealth(ctx); err != nil {
		log.Printf("[AI] Warning: Provider health check failed: %v", err)
	} else {
		log.Printf("[AI] Provider %s initialized successfully", provider.Name())
	}

	return NewClient(provider), nil
}

func (s *Service) IsEnabled() bool {
	return s.provider != nil
}

func (s *Service) SingleChat(ctx context.Context, prompt string) (string, error) {
	if _, err := s.requireProvider(); err != nil {
		return "", err
	}
	return s.provider.SingleChat(ctx, prompt)
}

func (s *Service) GetStatus() types.StatusDTO {
	provider, err := s.requireProvider()
	if err != nil {
		return types.StatusDTO{Status: "unavailable"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	status := "ready"
	if err := provider.CheckHealth(ctx); err != nil {
		status = "unavailable"
	}

	return types.StatusDTO{
		Provider: provider.Name(),
		IsLocal:  provider.IsLocal(),
		Status:   status,
	}
}

func (s *Service) Diagnose(
	ctx context.Context,
	statsProvider metrics.StatsProvider,
	workloadReg *workload.Registry,
	store storage.EventStore,
	processTree *proc.ProcessTree,
) (*types.DiagnosisResult, error) {
	provider, err := s.requireProvider()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	result := s.buildSnapshot(statsProvider, workloadReg, store, processTree)
	promptText, err := diagnostics.BuildPrompt(result.State)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompt: %w", err)
	}

	response, err := provider.SingleChat(ctx, promptText)
	if err != nil {
		return nil, fmt.Errorf("AI inference failed: %w", err)
	}

	return &types.DiagnosisResult{
		Analysis:        response,
		SnapshotSummary: diagnostics.SnapshotSummary(result.State),
		Provider:        provider.Name(),
		IsLocal:         provider.IsLocal(),
		DurationMs:      time.Since(startTime).Milliseconds(),
		Timestamp:       nowMs(),
	}, nil
}

func (s *Service) Chat(
	ctx context.Context,
	sessionID string,
	userMessage string,
	statsProvider metrics.StatsProvider,
	workloadReg *workload.Registry,
	store storage.EventStore,
	processTree *proc.ProcessTree,
) (*types.ChatResponse, error) {
	provider, err := s.requireProvider()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	conv := s.conversations.GetOrCreate(sessionID)
	history := s.conversations.GetMessages(sessionID)

	result := s.buildSnapshot(statsProvider, workloadReg, store, processTree)
	messages := chat.BuildMessages(history, result.State, userMessage, processTree, result.ProcessKeyToChain, result.ProcessNameToChain)

	response, err := provider.MultiChat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("AI chat failed: %w", err)
	}

	ts := nowMs()
	s.appendUserAndAssistant(sessionID, userMessage, response, ts)

	return &types.ChatResponse{
		Message:        response,
		SessionID:      sessionID,
		ContextSummary: diagnostics.SnapshotSummary(result.State),
		Provider:       provider.Name(),
		IsLocal:        provider.IsLocal(),
		DurationMs:     time.Since(startTime).Milliseconds(),
		Timestamp:      time.Now().UnixMilli(),
		MessageCount:   len(conv.Messages),
	}, nil
}

func (s *Service) ChatStream(
	ctx context.Context,
	sessionID string,
	userMessage string,
	statsProvider metrics.StatsProvider,
	workloadReg *workload.Registry,
	store storage.EventStore,
	processTree *proc.ProcessTree,
) (<-chan types.ChatStreamToken, error) {
	provider, err := s.requireProvider()
	if err != nil {
		return nil, err
	}

	s.conversations.GetOrCreate(sessionID)
	history := s.conversations.GetMessages(sessionID)

	result := s.buildSnapshot(statsProvider, workloadReg, store, processTree)
	messages := chat.BuildMessages(history, result.State, userMessage, processTree, result.ProcessKeyToChain, result.ProcessNameToChain)

	tokenChan, err := provider.MultiChatStream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("AI stream failed: %w", err)
	}

	outputChan := make(chan types.ChatStreamToken, 100)

	go func() {
		defer close(outputChan)

		var fullResponse string

		for token := range tokenChan {
			if token.Error != nil {
				outputChan <- types.ChatStreamToken{Error: token.Error.Error()}
				return
			}

			fullResponse += token.Content

			outputChan <- types.ChatStreamToken{
				Content:   token.Content,
				Done:      token.Done,
				SessionID: sessionID,
			}

			if token.Done {
				ts := nowMs()
				s.appendUserAndAssistant(sessionID, userMessage, fullResponse, ts)
			}
		}
	}()

	return outputChan, nil
}

func (s *Service) GetChatHistory(sessionID string) []types.Message {
	if s.conversations == nil {
		return nil
	}
	return s.conversations.GetMessages(sessionID)
}

func (s *Service) ClearChat(sessionID string) {
	if s.conversations != nil {
		s.conversations.Clear(sessionID)
	}
}

func (s *Service) AskAboutInsight(ctx context.Context, req *types.AskInsightRequest) (*types.AskInsightResponse, error) {
	provider, err := s.requireProvider()
	if err != nil {
		return nil, err
	}

	// Simple prompt construction for now. A more advanced version could use templates
	// and fetch related events for even deeper context.
	insightJSON, err := json.MarshalIndent(req.Insight, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize insight: %w", err)
	}

	prompt := fmt.Sprintf("As a security analyst AI, answer the user's question about the following security insight.\n\n"+
		"User Question: %s\n\n"+
		"Full Insight Context (JSON):\n%s\n\n"+
		"Provide a direct answer to the question, leveraging all the details in the insight context. "+
		"Explain your reasoning clearly. If the insight data is insufficient, state what additional information you would need.",
		req.Question, string(insightJSON))

	response, err := provider.SingleChat(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI inference failed: %w", err)
	}

	// For now, we'll just wrap the raw response. A more advanced implementation
	// would have the AI return a structured JSON object with confidence scores.
	return &types.AskInsightResponse{
		Answer:     response,
		Confidence: 0.8, // Placeholder confidence
	}, nil
}
