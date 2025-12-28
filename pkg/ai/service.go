package ai

import (
	"context"
	"fmt"
	"log"
	"time"

	"aegis/pkg/ai/chat"
	"aegis/pkg/ai/diagnostics"
	"aegis/pkg/ai/providers"
	"aegis/pkg/ai/snapshot"
	"aegis/pkg/ai/types"
	"aegis/pkg/config"
	"aegis/pkg/metrics"
	"aegis/pkg/proc"
	"aegis/pkg/storage"
	"aegis/pkg/workload"
)

type Service struct {
	provider      Provider
	conversations *chat.Store
}

func NewClient(p Provider) *Service {
	return &Service{
		provider:      p,
		conversations: chat.NewStore(),
	}
}

func NewService(opts config.AIOptions) (*Service, error) {
	var provider Provider

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
	if s.provider == nil {
		return "", fmt.Errorf("AI service is not available")
	}
	return s.provider.SingleChat(ctx, prompt)
}

func (s *Service) GetStatus() types.StatusDTO {
	if s.provider == nil {
		return types.StatusDTO{
			Status:   "unavailable",
			Provider: "",
			IsLocal:  false,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	status := "ready"
	if err := s.provider.CheckHealth(ctx); err != nil {
		status = "unavailable"
	}

	return types.StatusDTO{
		Provider: s.provider.Name(),
		IsLocal:  s.provider.IsLocal(),
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
	if s.provider == nil {
		return nil, fmt.Errorf("AI service is not available")
	}

	startTime := time.Now()

	result := snapshot.NewSnapshot(statsProvider, workloadReg, store, processTree).Build()
	promptText, err := diagnostics.BuildPrompt(result.State)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompt: %w", err)
	}

	response, err := s.provider.SingleChat(ctx, promptText)
	if err != nil {
		return nil, fmt.Errorf("AI inference failed: %w", err)
	}

	return &types.DiagnosisResult{
		Analysis:        response,
		SnapshotSummary: diagnostics.SnapshotSummary(result.State),
		Provider:        s.provider.Name(),
		IsLocal:         s.provider.IsLocal(),
		DurationMs:      time.Since(startTime).Milliseconds(),
		Timestamp:       time.Now().UnixMilli(),
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
	if s.provider == nil {
		return nil, fmt.Errorf("AI service is not available")
	}

	startTime := time.Now()

	conv := s.conversations.GetOrCreate(sessionID)
	history := s.conversations.GetMessages(sessionID)

	result := snapshot.NewSnapshot(statsProvider, workloadReg, store, processTree).Build()
	messages := chat.BuildMessages(history, result.State, userMessage, processTree, result.ProcessKeyToChain, result.ProcessNameToChain)

	response, err := s.provider.MultiChat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("AI chat failed: %w", err)
	}

	s.conversations.AddMessage(sessionID, types.Message{
		Role:      "user",
		Content:   userMessage,
		Timestamp: time.Now().UnixMilli(),
	})
	s.conversations.AddMessage(sessionID, types.Message{
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now().UnixMilli(),
	})

	return &types.ChatResponse{
		Message:        response,
		SessionID:      sessionID,
		ContextSummary: diagnostics.SnapshotSummary(result.State),
		Provider:       s.provider.Name(),
		IsLocal:        s.provider.IsLocal(),
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
	if s.provider == nil {
		return nil, fmt.Errorf("AI service is not available")
	}

	s.conversations.GetOrCreate(sessionID)
	history := s.conversations.GetMessages(sessionID)

	result := snapshot.NewSnapshot(statsProvider, workloadReg, store, processTree).Build()
	messages := chat.BuildMessages(history, result.State, userMessage, processTree, result.ProcessKeyToChain, result.ProcessNameToChain)

	tokenChan, err := s.provider.MultiChatStream(ctx, messages)
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
				s.conversations.AddMessage(sessionID, types.Message{
					Role:      "user",
					Content:   userMessage,
					Timestamp: time.Now().UnixMilli(),
				})
				s.conversations.AddMessage(sessionID, types.Message{
					Role:      "assistant",
					Content:   fullResponse,
					Timestamp: time.Now().UnixMilli(),
				})
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
