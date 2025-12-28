package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultProcessTreeMaxAge         = 30 * time.Minute // 30 minutes
	DefaultProcessTreeMaxSize        = 10000
	DefaultProcessTreeMaxChainLength = 50
	DefaultRingBufferSize            = 256 * 1024 // 256KB
)

type Options struct {
	BPFPath                   string        `yaml:"bpf_path"`
	RulesPath                 string        `yaml:"rules_path"`
	RingBufferSize            int           `yaml:"ring_buffer_size"`
	ProcessTreeMaxAge         time.Duration `yaml:"process_tree_max_age"`
	ProcessTreeMaxSize        int           `yaml:"process_tree_max_size"`
	ProcessTreeMaxChainLength int           `yaml:"process_tree_max_chain_length"`

	// Rule promotion configuration
	PromotionMinObservationMinutes int `yaml:"promotion_min_observation_minutes"`
	PromotionMinHits               int `yaml:"promotion_min_hits"`

	WebPort int `yaml:"-"`

	// AI configuration
	AI AIOptions `yaml:"ai"`
}

type AIOptions struct {
	Mode   string        `yaml:"mode"` // "ollama" or "openai"
	Ollama OllamaOptions `yaml:"ollama"`
	OpenAI OpenAIOptions `yaml:"openai"`

	// Optional Sentinel schedule overrides (durations as strings, e.g. "5m").
	SentinelTestingPromotion string `yaml:"sentinel_testing_promotion"`
	SentinelAnomaly          string `yaml:"sentinel_anomaly"`
	SentinelRuleOptimization string `yaml:"sentinel_rule_optimization"`
	SentinelDailyReport      string `yaml:"sentinel_daily_report"`
}

type OllamaOptions struct {
	Endpoint string `yaml:"endpoint"`
	Model    string `yaml:"model"`
	Timeout  int    `yaml:"timeout"`
}

type OpenAIOptions struct {
	Endpoint string `yaml:"endpoint"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
	Timeout  int    `yaml:"timeout"`
}

func ParseOptions() Options {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	configPath := filepath.Join(cwd, "config.yaml")

	opts := Options{
		BPFPath:                        filepath.Join(cwd, "bpf", "main.bpf.o"),
		RulesPath:                      filepath.Join(cwd, "rules.yaml"),
		ProcessTreeMaxAge:              DefaultProcessTreeMaxAge,
		ProcessTreeMaxSize:             DefaultProcessTreeMaxSize,
		ProcessTreeMaxChainLength:      DefaultProcessTreeMaxChainLength,
		RingBufferSize:                 DefaultRingBufferSize,
		PromotionMinObservationMinutes: 1440, // 24 hours
		PromotionMinHits:               100,
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: failed to read config file: %v\n", err)
			os.Exit(1)
		}
		return opts
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse config file: %v\n", err)
		os.Exit(1)
	}

	if v, ok := raw["bpf_path"].(string); ok && v != "" {
		opts.BPFPath = v
	}
	if v, ok := raw["rules_path"].(string); ok && v != "" {
		opts.RulesPath = v
	}
	if v, ok := raw["ring_buffer_size"].(int); ok && v > 0 {
		opts.RingBufferSize = v
	}
	if v, ok := raw["process_tree_max_size"].(int); ok && v > 0 {
		opts.ProcessTreeMaxSize = v
	}
	if v, ok := raw["process_tree_max_chain_length"].(int); ok && v > 0 {
		opts.ProcessTreeMaxChainLength = v
	}
	if v, ok := raw["process_tree_max_age"].(string); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			opts.ProcessTreeMaxAge = d
		}
	}

	// Rule promotion configuration
	if v, ok := raw["promotion_min_observation_minutes"].(int); ok && v > 0 {
		opts.PromotionMinObservationMinutes = v
	}
	if v, ok := raw["promotion_min_hits"].(int); ok && v > 0 {
		opts.PromotionMinHits = v
	}

	// AI configuration
	if aiRaw, ok := raw["ai"].(map[string]any); ok {
		if v, ok := aiRaw["mode"].(string); ok {
			opts.AI.Mode = v
		}
		if ollamaRaw, ok := aiRaw["ollama"].(map[string]any); ok {
			if v, ok := ollamaRaw["endpoint"].(string); ok {
				opts.AI.Ollama.Endpoint = v
			}
			if v, ok := ollamaRaw["model"].(string); ok {
				opts.AI.Ollama.Model = v
			}
			if v, ok := ollamaRaw["timeout"].(int); ok {
				opts.AI.Ollama.Timeout = v
			}
		}
		if openaiRaw, ok := aiRaw["openai"].(map[string]any); ok {
			if v, ok := openaiRaw["endpoint"].(string); ok {
				opts.AI.OpenAI.Endpoint = v
			}
			if v, ok := openaiRaw["api_key"].(string); ok {
				opts.AI.OpenAI.APIKey = v
			}
			if v, ok := openaiRaw["model"].(string); ok {
				opts.AI.OpenAI.Model = v
			}
			if v, ok := openaiRaw["timeout"].(int); ok {
				opts.AI.OpenAI.Timeout = v
			}
		}

		if v, ok := aiRaw["sentinel_testing_promotion"].(string); ok && v != "" {
			opts.AI.SentinelTestingPromotion = v
		}
		if v, ok := aiRaw["sentinel_anomaly"].(string); ok && v != "" {
			opts.AI.SentinelAnomaly = v
		}
		if v, ok := aiRaw["sentinel_rule_optimization"].(string); ok && v != "" {
			opts.AI.SentinelRuleOptimization = v
		}
		if v, ok := aiRaw["sentinel_daily_report"].(string); ok && v != "" {
			opts.AI.SentinelDailyReport = v
		}
	}

	// Set AI defaults if not configured
	if opts.AI.Mode == "" {
		opts.AI.Mode = "ollama"
	}
	if opts.AI.Ollama.Endpoint == "" {
		opts.AI.Ollama.Endpoint = "http://localhost:11434"
	}
	if opts.AI.Ollama.Model == "" {
		opts.AI.Ollama.Model = "qwen2.5-coder:1.5b"
	}
	if opts.AI.Ollama.Timeout == 0 {
		opts.AI.Ollama.Timeout = 60
	}
	if opts.AI.OpenAI.Endpoint == "" {
		opts.AI.OpenAI.Endpoint = "https://api.deepseek.com"
	}
	if opts.AI.OpenAI.Model == "" {
		opts.AI.OpenAI.Model = "deepseek-chat"
	}
	if opts.AI.OpenAI.Timeout == 0 {
		opts.AI.OpenAI.Timeout = 30
	}

	// Parse command line flags (override config file)
	flag.IntVar(&opts.WebPort, "port", 3000, "Port for web GUI (default: 3000)")
	flag.Parse()

	return opts
}
