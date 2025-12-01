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
	DefaultLogBufferSize             = 64 * 1024  // 64KB
	DefaultRingBufferSize            = 256 * 1024 // 256KB
	DefaultLearnDuration             = 5 * time.Minute
)

type Options struct {
	BPFPath                   string        `yaml:"bpf_path"`
	RulesPath                 string        `yaml:"rules_path"`
	LogFile                   string        `yaml:"log_file"`
	JSONLines                 bool          `yaml:"json_output"`
	RingBufferSize            int           `yaml:"ring_buffer_size"`
	ProcessTreeMaxAge         time.Duration `yaml:"process_tree_max_age"`
	ProcessTreeMaxSize        int           `yaml:"process_tree_max_size"`
	ProcessTreeMaxChainLength int           `yaml:"process_tree_max_chain_length"`
	LogBufferSize             int           `yaml:"log_buffer_size"`

	LearnMode     bool          `yaml:"learn_mode"`
	LearnDuration time.Duration `yaml:"learn_duration"`

	WebMode bool `yaml:"-"`
	WebPort int  `yaml:"-"`

	// AI configuration
	AI AIOptions `yaml:"ai"`
}

type AIOptions struct {
	Enabled bool          `yaml:"enabled"`
	Mode    string        `yaml:"mode"` // "ollama" or "openai"
	Ollama  OllamaOptions `yaml:"ollama"`
	OpenAI  OpenAIOptions `yaml:"openai"`
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
		BPFPath:                   filepath.Join(cwd, "bpf", "main.bpf.o"),
		RulesPath:                 filepath.Join(cwd, "rules.yaml"),
		LogFile:                   filepath.Join(cwd, "eulerguard.log"),
		ProcessTreeMaxAge:         DefaultProcessTreeMaxAge,
		ProcessTreeMaxSize:        DefaultProcessTreeMaxSize,
		ProcessTreeMaxChainLength: DefaultProcessTreeMaxChainLength,
		LogBufferSize:             DefaultLogBufferSize,
		RingBufferSize:            DefaultRingBufferSize,
		LearnMode:     false,
		LearnDuration: DefaultLearnDuration,
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: failed to read config file: %v\n", err)
			os.Exit(1)
		}
		return opts
	}

	var raw map[string]interface{}
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
	if v, ok := raw["log_file"].(string); ok && v != "" {
		opts.LogFile = v
	}
	if v, ok := raw["json_output"].(bool); ok {
		opts.JSONLines = v
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
	if v, ok := raw["log_buffer_size"].(int); ok && v > 0 {
		opts.LogBufferSize = v
	}
	if v, ok := raw["process_tree_max_age"].(string); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			opts.ProcessTreeMaxAge = d
		}
	}

	// Learn mode options
	if v, ok := raw["learn_mode"].(bool); ok {
		opts.LearnMode = v
	}
	if v, ok := raw["learn_duration"].(string); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			opts.LearnDuration = d
		}
	}

	// AI configuration
	if aiRaw, ok := raw["ai"].(map[string]interface{}); ok {
		if v, ok := aiRaw["enabled"].(bool); ok {
			opts.AI.Enabled = v
		}
		if v, ok := aiRaw["mode"].(string); ok {
			opts.AI.Mode = v
		}
		if ollamaRaw, ok := aiRaw["ollama"].(map[string]interface{}); ok {
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
		if openaiRaw, ok := aiRaw["openai"].(map[string]interface{}); ok {
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
	flag.BoolVar(&opts.WebMode, "web", false, "Launch web GUI (accessible via browser)")
	flag.IntVar(&opts.WebPort, "port", 3000, "Port for web GUI (default: 3000)")
	flag.BoolVar(&opts.LearnMode, "learn", opts.LearnMode, "Enable learning mode")
	flag.Parse()

	return opts
}
