package system

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

type ConfigSaver interface {
	Save(Settings) error
}

type ConfigApplier interface {
	ApplyConfig(oldCfg, newCfg Settings, hotReloadedFields []string) error
}

type SettingsService struct {
	mu      sync.RWMutex
	cfg     Settings
	saver   ConfigSaver
	applier ConfigApplier
}

type UpdateResult struct {
	Updated               bool     `json:"updated"`
	Config                Settings `json:"config"`
	HotReloadedFields     []string `json:"hot_reloaded_fields,omitempty"`
	RestartRequired       bool     `json:"restart_required"`
	RestartRequiredFields []string `json:"restart_required_fields,omitempty"`
}

func NewSettingsService(cfg Settings, saver ConfigSaver) *SettingsService {
	return &SettingsService{cfg: cfg, saver: saver}
}

func (s *SettingsService) SetApplier(applier ConfigApplier) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.applier = applier
}

func (s *SettingsService) Get() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *SettingsService) Update(cfg Settings) (UpdateResult, error) {
	if err := validateConfig(cfg); err != nil {
		return UpdateResult{}, err
	}

	s.mu.RLock()
	oldCfg := s.cfg
	applier := s.applier
	s.mu.RUnlock()

	hotReloadedFields, restartRequiredFields := classifyConfigFields(oldCfg, cfg)
	changedFields := append(append([]string(nil), hotReloadedFields...), restartRequiredFields...)
	if s.saver != nil {
		if err := s.saver.Save(cfg); err != nil {
			return UpdateResult{}, err
		}
	}
	if applier != nil && len(hotReloadedFields) > 0 {
		if err := applier.ApplyConfig(oldCfg, cfg, hotReloadedFields); err != nil {
			return UpdateResult{}, err
		}
	}

	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()

	return UpdateResult{
		Updated:               len(changedFields) > 0,
		Config:                cfg,
		HotReloadedFields:     hotReloadedFields,
		RestartRequired:       len(restartRequiredFields) > 0,
		RestartRequiredFields: restartRequiredFields,
	}, nil
}

func validateConfig(cfg Settings) error {
	switch {
	case cfg.Server.Port <= 0:
		return fmt.Errorf("server.port must be greater than 0")
	case cfg.Kernel.BPFPath == "":
		return fmt.Errorf("kernel.bpf_path must not be empty")
	case cfg.Kernel.RingBufferSize <= 0:
		return fmt.Errorf("kernel.ring_buffer_size must be greater than 0")
	case cfg.Telemetry.ProcessTreeMaxSize <= 0:
		return fmt.Errorf("telemetry.process_tree_max_size must be greater than 0")
	case cfg.Telemetry.ProcessTreeMaxChainLength <= 0:
		return fmt.Errorf("telemetry.process_tree_max_chain_length must be greater than 0")
	case cfg.Telemetry.RecentEventsCapacity <= 0:
		return fmt.Errorf("telemetry.recent_events_capacity must be greater than 0")
	case cfg.Telemetry.EventIndexSize <= 0:
		return fmt.Errorf("telemetry.event_index_size must be greater than 0")
	case cfg.Policy.RulesPath == "":
		return fmt.Errorf("policy.rules_path must not be empty")
	case cfg.Policy.PromotionMinObservationMinutes < 0:
		return fmt.Errorf("policy.promotion_min_observation_minutes must not be negative")
	case cfg.Policy.PromotionMinHits < 0:
		return fmt.Errorf("policy.promotion_min_hits must not be negative")
	case cfg.Analysis.Mode != "" && cfg.Analysis.Mode != "disabled" && cfg.Analysis.Mode != "ollama" && cfg.Analysis.Mode != "openai" && cfg.Analysis.Mode != "gemini":
		return fmt.Errorf("analysis.mode must be one of disabled, ollama, openai, or gemini")
	case cfg.Analysis.Ollama.Timeout < 0:
		return fmt.Errorf("analysis.ollama.timeout must not be negative")
	case cfg.Analysis.OpenAI.Timeout < 0:
		return fmt.Errorf("analysis.openai.timeout must not be negative")
	case cfg.Analysis.Gemini.Timeout < 0:
		return fmt.Errorf("analysis.gemini.timeout must not be negative")
	case !isValidDuration(cfg.Sentinel.TestingPromotion):
		return fmt.Errorf("sentinel.testing_promotion must be a valid duration")
	case !isValidDuration(cfg.Sentinel.Anomaly):
		return fmt.Errorf("sentinel.anomaly must be a valid duration")
	case !isValidDuration(cfg.Sentinel.RuleOptimization):
		return fmt.Errorf("sentinel.rule_optimization must be a valid duration")
	case !isValidDuration(cfg.Sentinel.DailyReport):
		return fmt.Errorf("sentinel.daily_report must be a valid duration")
	default:
		return nil
	}
}

func classifyConfigFields(oldCfg, newCfg Settings) ([]string, []string) {
	hotReloaded := make([]string, 0)
	restartRequired := make([]string, 0)

	appendIfChanged := func(path string, oldValue, newValue any) {
		if !reflect.DeepEqual(oldValue, newValue) {
			if isHotReloadableField(path) {
				hotReloaded = append(hotReloaded, path)
				return
			}
			restartRequired = append(restartRequired, path)
		}
	}

	appendIfChanged("server.port", oldCfg.Server.Port, newCfg.Server.Port)
	appendIfChanged("kernel.bpf_path", oldCfg.Kernel.BPFPath, newCfg.Kernel.BPFPath)
	appendIfChanged("kernel.ring_buffer_size", oldCfg.Kernel.RingBufferSize, newCfg.Kernel.RingBufferSize)
	appendIfChanged("telemetry.process_tree_max_age", oldCfg.Telemetry.ProcessTreeMaxAge, newCfg.Telemetry.ProcessTreeMaxAge)
	appendIfChanged("telemetry.process_tree_max_size", oldCfg.Telemetry.ProcessTreeMaxSize, newCfg.Telemetry.ProcessTreeMaxSize)
	appendIfChanged("telemetry.process_tree_max_chain_length", oldCfg.Telemetry.ProcessTreeMaxChainLength, newCfg.Telemetry.ProcessTreeMaxChainLength)
	appendIfChanged("telemetry.recent_events_capacity", oldCfg.Telemetry.RecentEventsCapacity, newCfg.Telemetry.RecentEventsCapacity)
	appendIfChanged("telemetry.event_index_size", oldCfg.Telemetry.EventIndexSize, newCfg.Telemetry.EventIndexSize)
	appendIfChanged("policy.rules_path", oldCfg.Policy.RulesPath, newCfg.Policy.RulesPath)
	appendIfChanged("policy.promotion_min_observation_minutes", oldCfg.Policy.PromotionMinObservationMinutes, newCfg.Policy.PromotionMinObservationMinutes)
	appendIfChanged("policy.promotion_min_hits", oldCfg.Policy.PromotionMinHits, newCfg.Policy.PromotionMinHits)
	appendIfChanged("analysis.mode", oldCfg.Analysis.Mode, newCfg.Analysis.Mode)
	appendIfChanged("analysis.ollama", oldCfg.Analysis.Ollama, newCfg.Analysis.Ollama)
	appendIfChanged("analysis.openai", oldCfg.Analysis.OpenAI, newCfg.Analysis.OpenAI)
	appendIfChanged("analysis.gemini", oldCfg.Analysis.Gemini, newCfg.Analysis.Gemini)
	appendIfChanged("sentinel.testing_promotion", oldCfg.Sentinel.TestingPromotion, newCfg.Sentinel.TestingPromotion)
	appendIfChanged("sentinel.anomaly", oldCfg.Sentinel.Anomaly, newCfg.Sentinel.Anomaly)
	appendIfChanged("sentinel.rule_optimization", oldCfg.Sentinel.RuleOptimization, newCfg.Sentinel.RuleOptimization)
	appendIfChanged("sentinel.daily_report", oldCfg.Sentinel.DailyReport, newCfg.Sentinel.DailyReport)

	return hotReloaded, restartRequired
}

func isHotReloadableField(path string) bool {
	switch path {
	case "policy.promotion_min_observation_minutes",
		"policy.promotion_min_hits",
		"analysis.mode",
		"analysis.ollama",
		"analysis.openai",
		"analysis.gemini",
		"sentinel.testing_promotion",
		"sentinel.anomaly",
		"sentinel.rule_optimization",
		"sentinel.daily_report":
		return true
	default:
		return false
	}
}

func isValidDuration(raw string) bool {
	if raw == "" {
		return true
	}
	_, err := time.ParseDuration(raw)
	return err == nil
}
