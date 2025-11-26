package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"gopkg.in/yaml.v3"
)

func LoadRules(filePath string) ([]Rule, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	var ruleSet RuleSet
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

	return ruleSet.Rules, nil
}

func SaveRules(filePath string, ruleList []Rule) error {
	ruleSet := RuleSet{
		Rules: ruleList,
	}

	dir := filepath.Dir(filePath)
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

	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	if syscall.Geteuid() == 0 {
		if stat, err := os.Stat(filepath.Dir(filePath)); err == nil {
			uid := int(stat.Sys().(*syscall.Stat_t).Uid)
			gid := int(stat.Sys().(*syscall.Stat_t).Gid)
			os.Chown(filePath, uid, gid)
		}
	}

	return nil
}

func MergeRules(existing []Rule, newRules []Rule) []Rule {
	existingSet := make(map[string]bool)
	for _, r := range existing {
		sig := ruleSignature(r)
		existingSet[sig] = true
	}

	result := make([]Rule, len(existing))
	copy(result, existing)

	for _, r := range newRules {
		sig := ruleSignature(r)
		if !existingSet[sig] {
			result = append(result, r)
			existingSet[sig] = true
		}
	}

	return result
}

func ruleSignature(r Rule) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d|%s",
		r.Match.ProcessName,
		r.Match.ParentName,
		r.Match.Filename,
		r.Match.FilePath,
		r.Match.DestIP,
		r.Match.DestPort,
		r.Action,
	)
}
