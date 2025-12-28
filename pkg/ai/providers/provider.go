package providers

import (
	"context"

	"aegis/pkg/ai/types"
)

type Provider interface {
	Name() string

	IsLocal() bool

	// SingleChat handles a single-turn prompt/response interaction.
	SingleChat(ctx context.Context, prompt string) (string, error)

	// MultiChat handles multi-turn conversations using a sequence of messages.
	MultiChat(ctx context.Context, messages []types.Message) (string, error)

	// MultiChatStream handles streaming multi-turn conversations.
	MultiChatStream(ctx context.Context, messages []types.Message) (<-chan StreamToken, error)

	CheckHealth(ctx context.Context) error
}

type StreamToken struct {
	Content string
	Done    bool
	Error   error
}


