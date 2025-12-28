package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"aegis/pkg/config"
	"aegis/pkg/server"

	"gopkg.in/yaml.v3"
)

func RegisterSettingsHandlers(mux *http.ServeMux, app *server.App) {
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" {
			// Return current settings
			opts := app.Options()
			settings := map[string]any{
				"ai": map[string]any{
					"mode": opts.AI.Mode,
					"ollama": map[string]any{
						"endpoint": opts.AI.Ollama.Endpoint,
						"model":    opts.AI.Ollama.Model,
						"timeout":  opts.AI.Ollama.Timeout,
					},
					"openai": map[string]any{
						"endpoint": opts.AI.OpenAI.Endpoint,
						"apiKey":   opts.AI.OpenAI.APIKey,
						"model":    opts.AI.OpenAI.Model,
						"timeout":  opts.AI.OpenAI.Timeout,
					},
				},
				"testing": map[string]any{},
				"promotion": map[string]any{
					"minObservationMinutes": opts.PromotionMinObservationMinutes,
					"minHits":               opts.PromotionMinHits,
				},
			}
			json.NewEncoder(w).Encode(settings)
			return
		}

		if r.Method == "PUT" || r.Method == "POST" {
			// Update settings
			var req map[string]any
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			// Update app.opts (in-memory)
			opts := app.Options()
			if aiRaw, ok := req["ai"].(map[string]any); ok {
				if v, ok := aiRaw["mode"].(string); ok && (v == "ollama" || v == "openai") {
					opts.AI.Mode = v
				}
				if ollamaRaw, ok := aiRaw["ollama"].(map[string]any); ok {
					if v, ok := ollamaRaw["endpoint"].(string); ok {
						opts.AI.Ollama.Endpoint = v
					}
					if v, ok := ollamaRaw["model"].(string); ok {
						opts.AI.Ollama.Model = v
					}
					if v, ok := ollamaRaw["timeout"].(float64); ok {
						opts.AI.Ollama.Timeout = int(v)
					}
				}
				if openaiRaw, ok := aiRaw["openai"].(map[string]any); ok {
					if v, ok := openaiRaw["endpoint"].(string); ok {
						opts.AI.OpenAI.Endpoint = v
					}
					if v, ok := openaiRaw["apiKey"].(string); ok {
						opts.AI.OpenAI.APIKey = v
					}
					if v, ok := openaiRaw["model"].(string); ok {
						opts.AI.OpenAI.Model = v
					}
					if v, ok := openaiRaw["timeout"].(float64); ok {
						opts.AI.OpenAI.Timeout = int(v)
					}
				}
			}

			// Save to config.yaml file
			if err := saveConfigToFile(opts); err != nil {
				http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}

// saveConfigToFile saves the options to config.yaml file
func saveConfigToFile(opts *config.Options) error {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	configPath := filepath.Join(cwd, "config.yaml")

	data, err := yaml.Marshal(opts)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
