package bootstrap

import (
	"fmt"

	"aegis/pkg/ai"
	"aegis/pkg/ai/providers"
	"aegis/pkg/config"
)

func NewClientFromConfig(opts config.AIOptions) (*ai.Service, error) {
	var provider providers.Provider

	switch opts.Mode {
	case "ollama":
		provider = providers.NewOllamaProvider(opts.Ollama)
	case "openai":
		provider = providers.NewOpenAIProvider(opts.OpenAI)
	default:
		return nil, fmt.Errorf("unknown AI mode: %s", opts.Mode)
	}

	return ai.NewClient(provider), nil
}

