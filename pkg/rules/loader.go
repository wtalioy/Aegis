package rules

import (
	"fmt"
	"os"

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

	return ruleSet.Rules, nil
}

