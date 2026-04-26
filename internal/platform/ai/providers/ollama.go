package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"aegis/internal/analysis/types"
	"aegis/internal/platform/config"
)

type ollamaOptions struct {
	Temperature float64 `json:"temperature"`
	NumPredict  int     `json:"num_predict"`
}

type ollamaGenerateRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt"`
	Stream  bool          `json:"stream"`
	Options ollamaOptions `json:"options"`
}

type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []map[string]string `json:"messages"`
	Stream   bool                `json:"stream"`
	Options  ollamaOptions       `json:"options"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
}

type ollamaChatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type ollamaStreamChunk struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

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
	req, err := newJSONRequest(ctx, http.MethodPost, o.endpoint+"/api/generate", ollamaGenerateRequest{
		Model:  o.model,
		Prompt: prompt,
		Stream: false,
		Options: ollamaOptions{
			Temperature: defaultPromptTemperature,
			NumPredict:  defaultMaxTokens,
		},
	})
	if err != nil {
		return "", err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", unexpectedStatus("ollama", resp.StatusCode)
	}

	var result ollamaGenerateResponse
	if err := decodeJSONResponse(resp, &result); err != nil {
		return "", err
	}

	return result.Response, nil
}

func (o *OllamaProvider) MultiChat(ctx context.Context, messages []types.Message) (string, error) {
	req, err := newJSONRequest(ctx, http.MethodPost, o.endpoint+"/api/chat", ollamaChatRequest{
		Model:    o.model,
		Messages: ToRoleContent(messages),
		Stream:   false,
		Options: ollamaOptions{
			Temperature: defaultChatTemperature,
			NumPredict:  defaultMaxTokens,
		},
	})
	if err != nil {
		return "", err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama chat request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", unexpectedStatus("ollama", resp.StatusCode)
	}

	var result ollamaChatResponse
	if err := decodeJSONResponse(resp, &result); err != nil {
		return "", err
	}

	return result.Message.Content, nil
}

func (o *OllamaProvider) MultiChatStream(ctx context.Context, messages []types.Message) (<-chan StreamToken, error) {
	req, err := newJSONRequest(ctx, http.MethodPost, o.endpoint+"/api/chat", ollamaChatRequest{
		Model:    o.model,
		Messages: ToRoleContent(messages),
		Stream:   true,
		Options: ollamaOptions{
			Temperature: defaultChatTemperature,
			NumPredict:  defaultMaxTokens,
		},
	})
	if err != nil {
		return nil, err
	}

	resp, err := o.streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama stream request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, unexpectedStatus("ollama", resp.StatusCode)
	}

	tokenChan := make(chan StreamToken, 100)

	// Read streaming response in goroutine.
	go func() {
		defer close(tokenChan)
		defer resp.Body.Close()

		scanner := newStreamScanner(resp)

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

			var chunk ollamaStreamChunk
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
		return unexpectedStatus("ollama", resp.StatusCode)
	}
	return nil
}
