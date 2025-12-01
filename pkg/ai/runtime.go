package ai

import (
	"context"
	"log"
	"os"
	"os/exec"
)

const (
	defaultOllamaModel    = "qwen2.5-coder:1.5b"
	defaultOllamaEndpoint = "http://localhost:11434"
	startScriptName       = "scripts/start_ollama.sh"
)

// EnsureOllamaRuntime makes sure the local Ollama daemon is running and the
// requested model is pulled before the web UI comes up. It is safe to call
// multiple times—the helper script is idempotent.
func EnsureOllamaRuntime(ctx context.Context, model, endpoint string) error {
	if model == "" {
		model = defaultOllamaModel
	}
	if endpoint == "" {
		endpoint = defaultOllamaEndpoint
	}

	log.Printf("[AI] Preparing Ollama runtime (%s @ %s)...", model, endpoint)

	cmd := exec.CommandContext(ctx, startScriptName, model, endpoint)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
