package providers

import (
	"context"

	"aegis/internal/analysis/types"
)

// StreamToken is a single streamed output token from a provider.
type StreamToken struct {
	Content string
	Done    bool
	Error   error
}

// Provider defines the interface for AI model backends (Ollama, OpenAI, etc).
type Provider interface {
	Name() string
	IsLocal() bool
	CheckHealth(ctx context.Context) error

	SingleChat(ctx context.Context, prompt string) (string, error)
	MultiChat(ctx context.Context, messages []types.Message) (string, error)
	MultiChatStream(ctx context.Context, messages []types.Message) (<-chan StreamToken, error)
}
