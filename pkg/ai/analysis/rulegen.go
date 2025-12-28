package analysis

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"aegis/pkg/ai/prompt"
	"aegis/pkg/ai/providers"
	"aegis/pkg/ai/types"
	"aegis/pkg/rules"
	"aegis/pkg/storage"
	"gopkg.in/yaml.v3"
)

func GenerateRule(ctx context.Context, p providers.Provider, req *types.RuleGenRequest, ruleEngine *rules.Engine, store storage.EventStore) (*types.RuleGenResponse, error) {
	if p == nil {
		return nil, fmt.Errorf("AI provider is not available")
	}

	examplesYAML := ""
	for _, ex := range req.Examples {
		yb, _ := yaml.Marshal(ex)
		examplesYAML += string(yb) + "\n---\n"
	}

	fullPrompt := prompt.BuildRuleGenPrompt(req, examplesYAML)
	response, err := p.SingleChat(ctx, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI inference failed: %w", err)
	}

	ruleYAML := extractYAMLFromResponse(response)
	if ruleYAML == "" {
		return nil, fmt.Errorf("failed to extract rule YAML from AI response")
	}

	var rule rules.Rule
	if err := yaml.Unmarshal([]byte(ruleYAML), &rule); err != nil {
		return nil, fmt.Errorf("failed to parse generated rule YAML: %w", err)
	}

	cleanRule := rules.CleanRuleForYAML(rule)
	yamlBytes, err := yaml.Marshal(cleanRule)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rule to YAML: %w", err)
	}

	reasoning, warnings := extractReasoningAndWarnings(response)

	resp := &types.RuleGenResponse{
		Rule:       rule,
		YAML:       string(yamlBytes),
		Reasoning:  reasoning,
		Confidence: 0.8,
		Warnings:   warnings,
	}

	return resp, nil
}

func extractYAMLFromResponse(text string) string {
	if text == "" {
		return ""
	}

	re := regexp.MustCompile("(?s)```(?i:(?:yaml|yml))?\\s*(.*?)```")
	if m := re.FindStringSubmatch(text); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}

	lines := strings.Split(text, "\n")
	start := -1
	for i, ln := range lines {
		l := strings.TrimSpace(ln)
		if strings.HasPrefix(l, "name:") || strings.HasPrefix(l, "match:") || strings.HasPrefix(l, "action:") {
			start = i
			break
		}
	}
	if start >= 0 {
		var b strings.Builder
		for i := start; i < len(lines); i++ {
			l := lines[i]
			trim := strings.TrimSpace(l)
			if strings.HasPrefix(trim, "Reasoning:") || strings.HasPrefix(trim, "Warnings:") || strings.HasPrefix(trim, "---") {
				break
			}
			b.WriteString(l)
			b.WriteString("\n")
		}
		return strings.TrimSpace(b.String())
	}

	trim := strings.TrimSpace(text)
	if strings.HasPrefix(trim, "name:") || strings.HasPrefix(trim, "match:") || strings.HasPrefix(trim, "action:") {
		return trim
	}

	return ""
}

func extractReasoningAndWarnings(text string) (string, []string) {
	reasoning := text
	warnings := []string{}
	return reasoning, warnings
}

