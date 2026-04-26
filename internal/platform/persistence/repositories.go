package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"aegis/internal/platform/config"
	"aegis/internal/policy"
	"aegis/internal/policy/rules"
	"gopkg.in/yaml.v3"
)

type RuleRepository struct {
	path string
}

func NewRuleRepository(path string) *RuleRepository {
	return &RuleRepository{path: path}
}

func (r *RuleRepository) Load() ([]policy.Rule, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	var ruleSet policy.RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return nil, fmt.Errorf("failed to parse rules YAML: %w", err)
	}

	if len(ruleSet.Rules) == 0 {
		return nil, fmt.Errorf("no rules found in file")
	}

	for i := range ruleSet.Rules {
		if ruleSet.Rules[i].Type == "" {
			ruleSet.Rules[i].Type = ruleSet.Rules[i].DeriveType()
		}
	}

	if errs := rules.ValidateRules(ruleSet.Rules); len(errs) > 0 {
		var b strings.Builder
		b.WriteString("rule validation failed:\n")
		for _, err := range errs {
			b.WriteString(" - ")
			b.WriteString(err.Error())
			b.WriteByte('\n')
		}
		return nil, fmt.Errorf("%s", strings.TrimSpace(b.String()))
	}

	return ruleSet.Rules, nil
}

func (r *RuleRepository) Save(ruleList []policy.Rule) error {
	cleanRules := make([]policy.Rule, len(ruleList))
	for i, rule := range ruleList {
		cleanRules[i] = rules.CleanRuleForYAML(rule)
	}

	ruleSet := policy.RuleSet{Rules: cleanRules}

	dir := filepath.Dir(r.path)
	tmpFile, err := os.CreateTemp(dir, ".rules-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	encoder := yaml.NewEncoder(tmpFile)
	encoder.SetIndent(2)
	if err := encoder.Encode(ruleSet); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to encode rules to YAML: %w", err)
	}
	if err := encoder.Close(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close encoder: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, r.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	if syscall.Geteuid() == 0 {
		if stat, err := os.Stat(dir); err == nil {
			uid := int(stat.Sys().(*syscall.Stat_t).Uid)
			gid := int(stat.Sys().(*syscall.Stat_t).Gid)
			_ = os.Chown(r.path, uid, gid)
		}
	}

	return nil
}

func (r *RuleRepository) Path() string {
	return r.path
}

type ConfigRepository struct {
	path string
}

func NewConfigRepository(path string) *ConfigRepository {
	return &ConfigRepository{path: path}
}

func (r *ConfigRepository) Save(cfg config.Config) error {
	return config.Save(r.path, cfg)
}

func (r *ConfigRepository) Path() string {
	return r.path
}
