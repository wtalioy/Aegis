package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"aegis/pkg/ai/types"
	"aegis/pkg/config"
)

type OllamaProvider struct {
	endpoint     string
	model        string
	client       *http.Client
	streamClient *http.Client
}

func NewOllamaProvider(opts config.OllamaOptions) *OllamaProvider {
	return &OllamaProvider{
		endpoint: opts.Endpoint,
		model:    opts.Model,
		client: &http.Client{
			Timeout: time.Duration(opts.Timeout) * time.Second,
		},
		streamClient: &http.Client{},
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

func (o *OllamaProvider) MultiChat(ctx context.Context, messages []types.Message) (string, error) {
	ollamaMessages := ToRoleContent(messages)

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

func (o *OllamaProvider) MultiChatStream(ctx context.Context, messages []types.Message) (<-chan StreamToken, error) {
	ollamaMessages := ToRoleContent(messages)

	reqBody := map[string]any{
		"model":    o.model,
		"messages": ollamaMessages,
		"stream":   true,
		"options": map[string]any{
			"temperature": 0.4,
			"num_predict": 2048,
		},
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST",
		o.endpoint+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama stream request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	tokenChan := make(chan StreamToken, 100)

	// Read streaming response in goroutine.
	go func() {
		defer close(tokenChan)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		// Increase buffer size for long lines.
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				tokenChan <- StreamToken{Error: ctx.Err()}
				return
			default:
			}

			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var chunk struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				Done bool `json:"done"`
			}

			if err := json.Unmarshal(line, &chunk); err != nil {
				tokenChan <- StreamToken{Error: fmt.Errorf("failed to parse chunk: %w", err)}
				return
			}

			tokenChan <- StreamToken{
				Content: chunk.Message.Content,
				Done:    chunk.Done,
			}

			if chunk.Done {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			tokenChan <- StreamToken{Error: fmt.Errorf("scanner error: %w", err)}
		}
	}()

	return tokenChan, nil
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


