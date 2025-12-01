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

var ollamaStartedByUs bool

func EnsureOllamaRuntime(ctx context.Context, model, endpoint string) error {
	if model == "" {
		model = defaultOllamaModel
	}
	if endpoint == "" {
		endpoint = defaultOllamaEndpoint
	}

	if exec.Command("pgrep", "-x", "ollama").Run() != nil {
		ollamaStartedByUs = true
	}

	log.Printf("[AI] Preparing Ollama runtime (%s @ %s)...", model, endpoint)

	cmd := exec.CommandContext(ctx, startScriptName, model, endpoint)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func StopOllamaRuntime() {
	if !ollamaStartedByUs {
		return
	}
	log.Println("[AI] Stopping Ollama...")
	_ = exec.Command("pkill", "-x", "ollama").Run()
}
