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
	DefaultProcessTreeMaxAge         = 30 * time.Minute
	DefaultProcessTreeMaxSize        = 10000
	DefaultProcessTreeMaxChainLength = 50
	DefaultRingBufferSize            = 256 * 1024
	DefaultRecentEventsCapacity      = 10000
	DefaultMaxAlerts                 = 100
	DefaultAlertDedupWindow          = 10 * time.Second
)

type AIOptions struct {
	Mode   string        `yaml:"mode"`
	Ollama OllamaOptions `yaml:"ollama"`
	OpenAI OpenAIOptions `yaml:"openai"`
	Gemini GeminiOptions `yaml:"gemini"`

	SentinelTestingPromotion string `yaml:"sentinel_testing_promotion"`
	SentinelAnomaly          string `yaml:"sentinel_anomaly"`
	SentinelRuleOptimization string `yaml:"sentinel_rule_optimization"`
	SentinelDailyReport      string `yaml:"sentinel_daily_report"`
}

type OllamaOptions = ProviderConfig
type OpenAIOptions = OpenAIProviderConfig
type GeminiOptions = OpenAIProviderConfig

type Config struct {
	Server    ServerConfig    `yaml:"server" json:"server"`
	Kernel    KernelConfig    `yaml:"kernel" json:"kernel"`
	Telemetry TelemetryConfig `yaml:"telemetry" json:"telemetry"`
	Policy    PolicyConfig    `yaml:"policy" json:"policy"`
	Analysis  AnalysisConfig  `yaml:"analysis" json:"analysis"`
	Sentinel  SentinelConfig  `yaml:"sentinel" json:"sentinel"`
}

type ServerConfig struct {
	Port int `yaml:"port" json:"port"`
}

type KernelConfig struct {
	BPFPath        string `yaml:"bpf_path" json:"bpf_path"`
	RingBufferSize int    `yaml:"ring_buffer_size" json:"ring_buffer_size"`
}

type TelemetryConfig struct {
	ProcessTreeMaxAge         string `yaml:"process_tree_max_age" json:"process_tree_max_age"`
	ProcessTreeMaxSize        int    `yaml:"process_tree_max_size" json:"process_tree_max_size"`
	ProcessTreeMaxChainLength int    `yaml:"process_tree_max_chain_length" json:"process_tree_max_chain_length"`
	RecentEventsCapacity      int    `yaml:"recent_events_capacity" json:"recent_events_capacity"`
	EventIndexSize            int    `yaml:"event_index_size" json:"event_index_size"`
}

type PolicyConfig struct {
	RulesPath                      string `yaml:"rules_path" json:"rules_path"`
	PromotionMinObservationMinutes int    `yaml:"promotion_min_observation_minutes" json:"promotion_min_observation_minutes"`
	PromotionMinHits               int    `yaml:"promotion_min_hits" json:"promotion_min_hits"`
}

type AnalysisConfig struct {
	Mode   string               `yaml:"mode" json:"mode"`
	Ollama ProviderConfig       `yaml:"ollama" json:"ollama"`
	OpenAI OpenAIProviderConfig `yaml:"openai" json:"openai"`
	Gemini OpenAIProviderConfig `yaml:"gemini" json:"gemini"`
}

type ProviderConfig struct {
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	Model    string `yaml:"model" json:"model"`
	Timeout  int    `yaml:"timeout" json:"timeout"`
}

type OpenAIProviderConfig struct {
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	APIKey   string `yaml:"api_key" json:"api_key"`
	Model    string `yaml:"model" json:"model"`
	Timeout  int    `yaml:"timeout" json:"timeout"`
}

type SentinelConfig struct {
	TestingPromotion string `yaml:"testing_promotion" json:"testing_promotion"`
	Anomaly          string `yaml:"anomaly" json:"anomaly"`
	RuleOptimization string `yaml:"rule_optimization" json:"rule_optimization"`
	DailyReport      string `yaml:"daily_report" json:"daily_report"`
}

func Default(cwd string) Config {
	return Config{
		Server: ServerConfig{
			Port: 3000,
		},
		Kernel: KernelConfig{
			BPFPath:        filepath.Join(cwd, "bpf", "main.bpf.o"),
			RingBufferSize: DefaultRingBufferSize,
		},
		Telemetry: TelemetryConfig{
			ProcessTreeMaxAge:         DefaultProcessTreeMaxAge.String(),
			ProcessTreeMaxSize:        DefaultProcessTreeMaxSize,
			ProcessTreeMaxChainLength: DefaultProcessTreeMaxChainLength,
			RecentEventsCapacity:      DefaultRecentEventsCapacity,
			EventIndexSize:            1000,
		},
		Policy: PolicyConfig{
			RulesPath:                      filepath.Join(cwd, "rules.yaml"),
			PromotionMinObservationMinutes: 1440,
			PromotionMinHits:               100,
		},
		Analysis: AnalysisConfig{
			Mode: "ollama",
			Ollama: ProviderConfig{
				Endpoint: "http://localhost:11434",
				Model:    "qwen2.5-coder:1.5b",
				Timeout:  60,
			},
			OpenAI: OpenAIProviderConfig{
				Endpoint: "https://api.deepseek.com",
				Model:    "deepseek-chat",
				Timeout:  30,
			},
			Gemini: OpenAIProviderConfig{
				Endpoint: "https://generativelanguage.googleapis.com",
				Model:    "gemini-3-flash-preview",
				Timeout:  30,
			},
		},
		Sentinel: SentinelConfig{
			TestingPromotion: "15m",
			Anomaly:          "5m",
			RuleOptimization: "1h",
			DailyReport:      "24h",
		},
	}
}

func Load(args []string) (Config, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	return LoadFrom(cwd, args)
}

func LoadFrom(cwd string, args []string) (Config, string, error) {
	cfg := Default(cwd)
	configPath := filepath.Join(cwd, "config.yaml")

	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, "", fmt.Errorf("parse config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return Config{}, "", fmt.Errorf("read config file: %w", err)
	}

	fs := flag.NewFlagSet("aegis", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.IntVar(&cfg.Server.Port, "port", cfg.Server.Port, "Port for web GUI")
	if err := fs.Parse(args); err != nil {
		return Config{}, "", err
	}

	cfg.Kernel.BPFPath = resolvePath(cwd, cfg.Kernel.BPFPath)
	cfg.Policy.RulesPath = resolvePath(cwd, cfg.Policy.RulesPath)

	return cfg, configPath, nil
}

func Save(path string, cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func (c Config) ProcessTreeMaxAge() time.Duration {
	if d, err := time.ParseDuration(c.Telemetry.ProcessTreeMaxAge); err == nil && d > 0 {
		return d
	}
	return DefaultProcessTreeMaxAge
}

func (c Config) AIOptions() AIOptions {
	return AIOptions{
		Mode: c.Analysis.Mode,
		Ollama: OllamaOptions{
			Endpoint: c.Analysis.Ollama.Endpoint,
			Model:    c.Analysis.Ollama.Model,
			Timeout:  c.Analysis.Ollama.Timeout,
		},
		OpenAI: OpenAIOptions{
			Endpoint: c.Analysis.OpenAI.Endpoint,
			APIKey:   c.Analysis.OpenAI.APIKey,
			Model:    c.Analysis.OpenAI.Model,
			Timeout:  c.Analysis.OpenAI.Timeout,
		},
		Gemini: GeminiOptions{
			Endpoint: c.Analysis.Gemini.Endpoint,
			APIKey:   c.Analysis.Gemini.APIKey,
			Model:    c.Analysis.Gemini.Model,
			Timeout:  c.Analysis.Gemini.Timeout,
		},
		SentinelTestingPromotion: c.Sentinel.TestingPromotion,
		SentinelAnomaly:          c.Sentinel.Anomaly,
		SentinelRuleOptimization: c.Sentinel.RuleOptimization,
		SentinelDailyReport:      c.Sentinel.DailyReport,
	}
}

func resolvePath(cwd, path string) string {
	if path == "" {
		return path
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(cwd, path)
}
