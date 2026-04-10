package fakes

import (
	"sync"

	"aegis/internal/policy"
)

type RuleRepository struct {
	mu    sync.Mutex
	Rules []policy.Rule
}

func NewRuleRepository(ruleList []policy.Rule) *RuleRepository {
	copyRules := append([]policy.Rule(nil), ruleList...)
	return &RuleRepository{Rules: copyRules}
}

func (r *RuleRepository) Load() ([]policy.Rule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]policy.Rule(nil), r.Rules...), nil
}

func (r *RuleRepository) Save(ruleList []policy.Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Rules = append([]policy.Rule(nil), ruleList...)
	return nil
}

type KernelSync struct {
	mu       sync.Mutex
	SyncCall int
	Rules    [][]policy.Rule
}

func (k *KernelSync) SyncRules(ruleList []policy.Rule) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.SyncCall++
	k.Rules = append(k.Rules, append([]policy.Rule(nil), ruleList...))
	return nil
}
