package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"eulerguard/pkg/config"
)

type OllamaProvider struct {
	endpoint string
	model    string
	client   *http.Client
}

func NewOllamaProvider(opts config.OllamaOptions) *OllamaProvider {
	return &OllamaProvider{
		endpoint: opts.Endpoint,
		model:    opts.Model,
		client: &http.Client{
			Timeout: time.Duration(opts.Timeout) * time.Second,
		},
	}
}

func (o *OllamaProvider) Name() string  { return "Ollama" }
func (o *OllamaProvider) IsLocal() bool { return true }

func (o *OllamaProvider) SingleChat(ctx context.Context, prompt string) (string, error) {
	reqBody := map[string]any{
		"model":  o.model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]any{
			"temperature": 0.3,
			"num_predict": 2048,
		},
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST",
		o.endpoint+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Response, nil
}

func (o *OllamaProvider) MultiChat(ctx context.Context, messages []Message) (string, error) {
	ollamaMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	reqBody := map[string]any{
		"model":    o.model,
		"messages": ollamaMessages,
		"stream":   false,
		"options": map[string]any{
			"temperature": 0.4,
			"num_predict": 2048,
		},
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST",
		o.endpoint+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama chat request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Message.Content, nil
}

func (o *OllamaProvider) CheckHealth(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", o.endpoint+"/api/tags", nil)
	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama not reachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}
	return nil
}
