package ai

import (
	"context"
	"fmt"
	"log"
	"time"

	"eulerguard/pkg/config"
	"eulerguard/pkg/types"
	"eulerguard/pkg/workload"
)

type Service struct {
	provider      Provider
	conversations *ConversationStore
	enabled       bool
}

func NewService(opts config.AIOptions) (*Service, error) {
	if !opts.Enabled {
		return &Service{enabled: false}, nil
	}

	var provider Provider

	switch opts.Mode {
	case "ollama":
		provider = NewOllamaProvider(opts.Ollama)
	case "openai":
		provider = NewOpenAIProvider(opts.OpenAI)
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

	return &Service{
		provider:      provider,
		conversations: NewConversationStore(),
		enabled:       true,
	}, nil
}

func (s *Service) IsEnabled() bool {
	return s.enabled && s.provider != nil
}

func (s *Service) GetStatus() StatusDTO {
	if !s.enabled {
		return StatusDTO{
			Enabled:  false,
			Provider: "",
			IsLocal:  false,
			Status:   "disabled",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	status := "ready"
	if err := s.provider.CheckHealth(ctx); err != nil {
		status = "unavailable"
	}

	return StatusDTO{
		Enabled:  true,
		Provider: s.provider.Name(),
		IsLocal:  s.provider.IsLocal(),
		Status:   status,
	}
}

func (s *Service) Diagnose(
	ctx context.Context,
	stats types.StatsProvider,
	workloadReg *workload.Registry,
	procTreeSize int,
	userQuery string,
) (*DiagnosisResult, error) {
	if !s.enabled {
		return nil, fmt.Errorf("AI diagnosis is not enabled")
	}

	startTime := time.Now()

	snapshot := BuildSnapshot(stats, workloadReg, procTreeSize)
	prompt, err := GeneratePrompt(snapshot, userQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompt: %w", err)
	}

	response, err := s.provider.SingleChat(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI inference failed: %w", err)
	}

	return &DiagnosisResult{
		Analysis:        response,
		SnapshotSummary: FormatSnapshotSummary(snapshot),
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
	stats types.StatsProvider,
	workloadReg *workload.Registry,
	procTreeSize int,
) (*ChatResponse, error) {
	if !s.enabled {
		return nil, fmt.Errorf("AI chat is not enabled")
	}

	startTime := time.Now()

	conv := s.conversations.GetOrCreate(sessionID)
	history := s.conversations.GetMessages(sessionID)

	snapshot := BuildSnapshot(stats, workloadReg, procTreeSize)
	messages := BuildChatMessages(history, snapshot, userMessage)

	response, err := s.provider.MultiChat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("AI chat failed: %w", err)
	}

	s.conversations.AddMessage(sessionID, Message{
		Role:      "user",
		Content:   userMessage,
		Timestamp: time.Now().UnixMilli(),
	})
	s.conversations.AddMessage(sessionID, Message{
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now().UnixMilli(),
	})

	return &ChatResponse{
		Message:        response,
		SessionID:      sessionID,
		ContextSummary: FormatSnapshotSummary(snapshot),
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
	stats types.StatsProvider,
	workloadReg *workload.Registry,
	procTreeSize int,
) (<-chan ChatStreamToken, error) {
	if !s.enabled {
		return nil, fmt.Errorf("AI chat is not enabled")
	}

	s.conversations.GetOrCreate(sessionID)
	history := s.conversations.GetMessages(sessionID)

	snapshot := BuildSnapshot(stats, workloadReg, procTreeSize)
	messages := BuildChatMessages(history, snapshot, userMessage)

	tokenChan, err := s.provider.MultiChatStream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("AI stream failed: %w", err)
	}

	outputChan := make(chan ChatStreamToken, 100)

	go func() {
		defer close(outputChan)

		var fullResponse string

		for token := range tokenChan {
			if token.Error != nil {
				outputChan <- ChatStreamToken{Error: token.Error.Error()}
				return
			}

			fullResponse += token.Content

			outputChan <- ChatStreamToken{
				Content:   token.Content,
				Done:      token.Done,
				SessionID: sessionID,
			}

			if token.Done {
				s.conversations.AddMessage(sessionID, Message{
					Role:      "user",
					Content:   userMessage,
					Timestamp: time.Now().UnixMilli(),
				})
				s.conversations.AddMessage(sessionID, Message{
					Role:      "assistant",
					Content:   fullResponse,
					Timestamp: time.Now().UnixMilli(),
				})
			}
		}
	}()

	return outputChan, nil
}

func (s *Service) GetChatHistory(sessionID string) []Message {
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
