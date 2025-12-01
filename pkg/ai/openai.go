package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"eulerguard/pkg/config"
)

type OpenAIProvider struct {
	endpoint     string
	apiKey       string
	model        string
	client       *http.Client
	streamClient *http.Client
}

func NewOpenAIProvider(opts config.OpenAIOptions) *OpenAIProvider {
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("EULERGUARD_AI_API_KEY")
	}

	return &OpenAIProvider{
		endpoint: opts.Endpoint,
		apiKey:   apiKey,
		model:    opts.Model,
		client: &http.Client{
			Timeout: time.Duration(opts.Timeout) * time.Second,
		},
		streamClient: &http.Client{},
	}
}

func (o *OpenAIProvider) Name() string  { return "Cloud AI" }
func (o *OpenAIProvider) IsLocal() bool { return false }

func (o *OpenAIProvider) SingleChat(ctx context.Context, prompt string) (string, error) {
	messages := []Message{
		{Role: "system", Content: DiagnosisSystemPrompt},
		{Role: "user", Content: prompt},
	}
	return o.MultiChat(ctx, messages)
}

func (o *OpenAIProvider) MultiChat(ctx context.Context, messages []Message) (string, error) {
	openaiMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	reqBody := map[string]any{
		"model":       o.model,
		"messages":    openaiMessages,
		"temperature": 0.4,
		"max_tokens":  2048,
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST",
		o.endpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return result.Choices[0].Message.Content, nil
}

func (o *OpenAIProvider) CheckHealth(ctx context.Context) error {
	if o.apiKey == "" {
		return fmt.Errorf("API key not configured")
	}
	return nil
}

func (o *OpenAIProvider) MultiChatStream(ctx context.Context, messages []Message) (<-chan StreamToken, error) {
	openaiMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	reqBody := map[string]any{
		"model":       o.model,
		"messages":    openaiMessages,
		"temperature": 0.4,
		"max_tokens":  2048,
		"stream":      true,
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST",
		o.endpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API stream request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("OpenAI returned status %d", resp.StatusCode)
	}

	tokenChan := make(chan StreamToken, 100)

	go func() {
		defer close(tokenChan)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				tokenChan <- StreamToken{Done: true}
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				tokenChan <- StreamToken{Error: fmt.Errorf("failed to parse stream chunk: %w", err)}
				return
			}

			for _, choice := range chunk.Choices {
				if choice.Delta.Content != "" {
					tokenChan <- StreamToken{Content: choice.Delta.Content}
				}
				if choice.FinishReason == "stop" {
					tokenChan <- StreamToken{Done: true}
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			tokenChan <- StreamToken{Error: fmt.Errorf("stream read error: %w", err)}
		} else {
			tokenChan <- StreamToken{Done: true}
		}
	}()

	return tokenChan, nil
}
