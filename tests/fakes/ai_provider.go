package fakes

import (
	"context"
	"sync"

	"aegis/internal/analysis/types"
	"aegis/internal/platform/ai/providers"
)

type AIProvider struct {
	mu              sync.Mutex
	NameValue       string
	Local           bool
	HealthErr       error
	SingleResponse  string
	MultiResponse   string
	StreamResponses []string
	Prompts         []string
	Messages        [][]types.Message
}

func NewAIProvider() *AIProvider {
	return &AIProvider{
		NameValue:      "fake-ai",
		Local:          true,
		SingleResponse: "ok",
		MultiResponse:  "ok",
	}
}

func (p *AIProvider) Name() string {
	return p.NameValue
}

func (p *AIProvider) IsLocal() bool {
	return p.Local
}

func (p *AIProvider) CheckHealth(ctx context.Context) error {
	return p.HealthErr
}

func (p *AIProvider) SingleChat(ctx context.Context, prompt string) (string, error) {
	p.mu.Lock()
	p.Prompts = append(p.Prompts, prompt)
	p.mu.Unlock()
	return p.SingleResponse, nil
}

func (p *AIProvider) MultiChat(ctx context.Context, messages []types.Message) (string, error) {
	p.mu.Lock()
	p.Messages = append(p.Messages, append([]types.Message(nil), messages...))
	p.mu.Unlock()
	return p.MultiResponse, nil
}

func (p *AIProvider) MultiChatStream(ctx context.Context, messages []types.Message) (<-chan providers.StreamToken, error) {
	p.mu.Lock()
	p.Messages = append(p.Messages, append([]types.Message(nil), messages...))
	p.mu.Unlock()

	ch := make(chan providers.StreamToken, len(p.StreamResponses)+1)
	go func() {
		defer close(ch)
		for _, token := range p.StreamResponses {
			ch <- providers.StreamToken{Content: token}
		}
		ch <- providers.StreamToken{Done: true}
	}()
	return ch, nil
}
