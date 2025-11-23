package config

import (
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

	return opts
}
