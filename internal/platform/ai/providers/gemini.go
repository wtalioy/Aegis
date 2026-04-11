package providers

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

	"aegis/internal/analysis/types"
	"aegis/internal/platform/ai/prompt"
	"aegis/internal/platform/config"
)

type GeminiProvider struct {
	endpoint     string
	apiKey       string
	model        string
	client       *http.Client
	streamClient *http.Client
}

type geminiPart struct {
	Text string `json:"text"`
}

func NewGeminiProvider(opts config.GeminiOptions) *GeminiProvider {
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("AEGIS_AI_API_KEY")
	}

	endpoint := opts.Endpoint
	if endpoint == "" {
		endpoint = "https://generativelanguage.googleapis.com"
	}

	return &GeminiProvider{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    opts.Model,
		client: &http.Client{
			Timeout: time.Duration(opts.Timeout) * time.Second,
		},
		streamClient: &http.Client{},
	}
}

func (g *GeminiProvider) Name() string  { return "Gemini" }
func (g *GeminiProvider) IsLocal() bool { return false }

func (g *GeminiProvider) SingleChat(ctx context.Context, userPrompt string) (string, error) {
	return g.MultiChat(ctx, []types.Message{
		{Role: "system", Content: prompt.DiagnosisSystemPrompt},
		{Role: "user", Content: userPrompt},
	})
}

func (g *GeminiProvider) MultiChat(ctx context.Context, messages []types.Message) (string, error) {
	body, _ := json.Marshal(g.requestBody(messages))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.modelURL("generateContent", false), bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var result struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.Error.Message != "" {
			return "", fmt.Errorf("gemini API error: %s", result.Error.Message)
		}
		return "", fmt.Errorf("gemini returned status %d", resp.StatusCode)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []geminiPart `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error.Message != "" {
		return "", fmt.Errorf("gemini API error: %s", result.Error.Message)
	}
	if len(result.Candidates) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	return partsText(result.Candidates[0].Content.Parts), nil
}

func (g *GeminiProvider) CheckHealth(ctx context.Context) error {
	if g.apiKey == "" {
		return fmt.Errorf("API key not configured")
	}
	return nil
}

func (g *GeminiProvider) MultiChatStream(ctx context.Context, messages []types.Message) (<-chan StreamToken, error) {
	body, _ := json.Marshal(g.requestBody(messages))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.modelURL("streamGenerateContent", true), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini stream request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		response, fallbackErr := g.MultiChat(ctx, messages)
		if fallbackErr != nil {
			return nil, fallbackErr
		}

		tokenChan := make(chan StreamToken, 2)
		go func() {
			defer close(tokenChan)
			if response != "" {
				tokenChan <- StreamToken{Content: response}
			}
			tokenChan <- StreamToken{Done: true}
		}()
		return tokenChan, nil
	}

	tokenChan := make(chan StreamToken, 100)

	go func() {
		defer close(tokenChan)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)

		var eventData strings.Builder

		flushEvent := func() bool {
			data := strings.TrimSpace(eventData.String())
			eventData.Reset()
			if data == "" {
				return false
			}
			if data == "[DONE]" {
				tokenChan <- StreamToken{Done: true}
				return true
			}

			var chunk struct {
				Candidates []struct {
					Content struct {
						Parts []geminiPart `json:"parts"`
					} `json:"content"`
					FinishReason string `json:"finishReason"`
				} `json:"candidates"`
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				tokenChan <- StreamToken{Error: fmt.Errorf("failed to parse Gemini stream chunk: %w", err)}
				return true
			}
			if chunk.Error.Message != "" {
				tokenChan <- StreamToken{Error: fmt.Errorf("gemini API error: %s", chunk.Error.Message)}
				return true
			}

			for _, candidate := range chunk.Candidates {
				if content := partsText(candidate.Content.Parts); content != "" {
					tokenChan <- StreamToken{Content: content}
				}
				if strings.EqualFold(candidate.FinishReason, "stop") {
					tokenChan <- StreamToken{Done: true}
					return true
				}
			}

			return false
		}

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				tokenChan <- StreamToken{Error: ctx.Err()}
				return
			default:
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				if flushEvent() {
					return
				}
				continue
			}
			if strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
				continue
			}

			if eventData.Len() > 0 {
				eventData.WriteByte('\n')
			}
			eventData.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}

		if flushEvent() {
			return
		}
		if err := scanner.Err(); err != nil {
			tokenChan <- StreamToken{Error: fmt.Errorf("stream read error: %w", err)}
			return
		}

		tokenChan <- StreamToken{Done: true}
	}()

	return tokenChan, nil
}

func (g *GeminiProvider) requestBody(messages []types.Message) map[string]any {
	systemParts := make([]string, 0)
	contents := make([]map[string]any, 0, len(messages))

	for _, msg := range messages {
		text := strings.TrimSpace(msg.Content)
		if text == "" {
			continue
		}

		switch msg.Role {
		case "system":
			systemParts = append(systemParts, text)
		case "assistant":
			contents = append(contents, map[string]any{
				"role": "model",
				"parts": []map[string]string{
					{"text": text},
				},
			})
		default:
			contents = append(contents, map[string]any{
				"role": "user",
				"parts": []map[string]string{
					{"text": text},
				},
			})
		}
	}

	body := map[string]any{
		"contents": contents,
	}
	if len(systemParts) > 0 {
		body["systemInstruction"] = map[string]any{
			"parts": []map[string]string{
				{"text": strings.Join(systemParts, "\n\n")},
			},
		}
	}
	if len(contents) == 0 {
		body["contents"] = []map[string]any{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": "Continue."},
				},
			},
		}
	}

	return body
}

func (g *GeminiProvider) modelURL(operation string, sse bool) string {
	base := strings.TrimRight(g.endpoint, "/")

	switch {
	case strings.Contains(base, ":generateContent"):
		base = strings.Replace(base, ":generateContent", ":"+operation, 1)
	case strings.Contains(base, ":streamGenerateContent"):
		base = strings.Replace(base, ":streamGenerateContent", ":"+operation, 1)
	case strings.Contains(base, "/models/"):
		base = base + ":" + operation
	case strings.HasSuffix(base, "/v1"), strings.HasSuffix(base, "/v1beta"):
		base = fmt.Sprintf("%s/models/%s:%s", base, g.model, operation)
	default:
		base = fmt.Sprintf("%s/v1beta/models/%s:%s", base, g.model, operation)
	}

	if sse {
		if strings.Contains(base, "alt=") {
			return base
		}
		if strings.Contains(base, "?") {
			return base + "&alt=sse"
		}
		return base + "?alt=sse"
	}

	return base
}

func partsText(parts []geminiPart) string {
	var builder strings.Builder
	for _, part := range parts {
		builder.WriteString(part.Text)
	}
	return builder.String()
}
