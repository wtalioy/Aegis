package ai

import "context"

type Provider interface {
	Name() string

	IsLocal() bool

	// single prompt mode
	SingleChat(ctx context.Context, prompt string) (string, error)

	// multi-turn conversations mode
	MultiChat(ctx context.Context, messages []Message) (string, error)

	// streaming multi-turn conversations mode
	MultiChatStream(ctx context.Context, messages []Message) (<-chan StreamToken, error)

	CheckHealth(ctx context.Context) error
}

type StreamToken struct {
	Content string
	Done    bool
	Error   error
}

type ChatStreamToken struct {
	Content   string `json:"content"`
	Done      bool   `json:"done"`
	SessionID string `json:"sessionId,omitempty"`
	Error     string `json:"error,omitempty"`
}

type Message struct {
	Role      string `json:"role"` // "user", "assistant", "system"
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type DiagnosisResult struct {
	Analysis        string `json:"analysis"`
	SnapshotSummary string `json:"snapshotSummary"`
	Provider        string `json:"provider"`
	IsLocal         bool   `json:"isLocal"`
	DurationMs      int64  `json:"durationMs"`
	Timestamp       int64  `json:"timestamp"`
}

type ChatResponse struct {
	Message        string `json:"message"`
	SessionID      string `json:"sessionId"`
	ContextSummary string `json:"contextSummary"`
	Provider       string `json:"provider"`
	IsLocal        bool   `json:"isLocal"`
	DurationMs     int64  `json:"durationMs"`
	Timestamp      int64  `json:"timestamp"`
	MessageCount   int    `json:"messageCount"`
}

type StatusDTO struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	IsLocal  bool   `json:"isLocal"`
	Status   string `json:"status"` // "ready", "unavailable", "disabled"
}
