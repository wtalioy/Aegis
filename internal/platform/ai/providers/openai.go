package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"aegis/internal/analysis/types"
	"aegis/internal/platform/ai/prompt"
	"aegis/internal/platform/config"
)

type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	Temperature float64             `json:"temperature"`
	MaxTokens   int                 `json:"max_tokens"`
	Stream      bool                `json:"stream,omitempty"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

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
		apiKey = os.Getenv("AEGIS_AI_API_KEY")
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

func (o *OpenAIProvider) SingleChat(ctx context.Context, userPrompt string) (string, error) {
	messages := []types.Message{
		{Role: "system", Content: prompt.DiagnosisSystemPrompt},
		{Role: "user", Content: userPrompt},
	}
	return o.MultiChat(ctx, messages)
}

func (o *OpenAIProvider) MultiChat(ctx context.Context, messages []types.Message) (string, error) {
	req, err := newJSONRequest(ctx, http.MethodPost, o.endpoint+"/v1/chat/completions", openAIChatRequest{
		Model:       o.model,
		Messages:    ToRoleContent(messages),
		Temperature: defaultChatTemperature,
		MaxTokens:   defaultMaxTokens,
	})
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", unexpectedStatus("OpenAI", resp.StatusCode)
	}

	var result openAIChatResponse
	if err := decodeJSONResponse(resp, &result); err != nil {
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

func (o *OpenAIProvider) MultiChatStream(ctx context.Context, messages []types.Message) (<-chan StreamToken, error) {
	req, err := newJSONRequest(ctx, http.MethodPost, o.endpoint+"/v1/chat/completions", openAIChatRequest{
		Model:       o.model,
		Messages:    ToRoleContent(messages),
		Temperature: defaultChatTemperature,
		MaxTokens:   defaultMaxTokens,
		Stream:      true,
	})
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API stream request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, unexpectedStatus("OpenAI", resp.StatusCode)
	}

	tokenChan := make(chan StreamToken, 100)

	go func() {
		defer close(tokenChan)
		defer resp.Body.Close()

		scanner := newStreamScanner(resp)

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

			var chunk openAIStreamChunk
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
