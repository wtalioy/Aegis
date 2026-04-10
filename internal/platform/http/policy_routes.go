package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"aegis/internal/policy"

	"gopkg.in/yaml.v3"
)

func registerPolicyRoutes(mux *http.ServeMux, deps Dependencies) {
	registerAliases(mux, []string{"/api/v1/policies/testing"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		items := deps.Policy.TestingRules()
		payload := make([]map[string]any, 0, len(items))
		for _, item := range items {
			payload = append(payload, map[string]any{
				"rule":       policyRuleDTO(item.Rule),
				"validation": item.Validation,
				"stats":      item.Stats,
			})
		}
		writeJSON(w, http.StatusOK, payload)
	})

	registerAliasesWithPrefix(mux, []string{"/api/v1/policies/validation/"}, func(w http.ResponseWriter, r *http.Request, name string) {
		setCORS(w)
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		readiness, stats, err := deps.Policy.Validation(name)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		rule, _ := deps.Policy.Get(name)
		writeJSON(w, http.StatusOK, map[string]any{
			"rule":       policyRuleDTOValue(rule),
			"validation": readiness,
			"stats":      stats,
		})
	})

	registerAliases(mux, []string{"/api/v1/policies"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		switch r.Method {
		case http.MethodGet:
			ruleList := deps.Policy.List()
			payload := make([]map[string]any, 0, len(ruleList))
			for _, rule := range ruleList {
				payload = append(payload, policyRuleDTO(rule))
			}
			writeJSON(w, http.StatusOK, payload)
		case http.MethodPost:
			var req struct {
				Rule map[string]any `json:"rule"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			rule, err := decodePolicyRule(req.Rule)
			if err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			created, err := deps.Policy.Create(rule)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"success": true, "rule": policyRuleDTO(created)})
		case http.MethodOptions:
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		default:
			methodNotAllowed(w)
		}
	})

	registerAliasesWithPrefix(mux, []string{"/api/v1/policies/"}, func(w http.ResponseWriter, r *http.Request, suffix string) {
		setCORS(w)
		if suffix == "" || suffix == "testing" || strings.HasPrefix(suffix, "validation/") {
			http.NotFound(w, r)
			return
		}

		if strings.HasSuffix(suffix, "/promote") {
			name := strings.TrimSuffix(suffix, "/promote")
			if r.Method != http.MethodPost {
				methodNotAllowed(w)
				return
			}
			if err := deps.Policy.Promote(name); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"success": true})
			return
		}

		name := suffix
		switch r.Method {
		case http.MethodGet:
			rule, ok := deps.Policy.Get(name)
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, http.StatusOK, policyRuleDTOValue(rule))
		case http.MethodPut:
			var req struct {
				Rule map[string]any `json:"rule"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			rule, err := decodePolicyRule(req.Rule)
			if err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			updated, err := deps.Policy.Update(name, rule)
			if err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"success": true, "rule": policyRuleDTO(updated)})
		case http.MethodDelete:
			if err := deps.Policy.Delete(name); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"success": true})
		case http.MethodOptions:
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		default:
			methodNotAllowed(w)
		}
	})
}

func policyRuleDTO(rule policy.Rule) map[string]any {
	cleanRule := policy.CleanRuleForYAML(rule)
	yamlBytes, _ := yaml.Marshal(cleanRule)

	return map[string]any{
		"name":        rule.Name,
		"description": rule.Description,
		"severity":    rule.Severity,
		"action":      string(rule.Action),
		"type":        string(rule.DeriveType()),
		"state":       string(rule.State),
		"match":       policyMatchMap(rule.Match),
		"yaml":        string(yamlBytes),
		"created_at":  rule.CreatedAt,
		"deployed_at": rule.DeployedAt,
		"promoted_at": rule.PromotedAt,
	}
}

func policyRuleDTOValue(rule *policy.Rule) map[string]any {
	if rule == nil {
		return nil
	}
	return policyRuleDTO(*rule)
}

func policyMatchMap(match policy.MatchCondition) map[string]any {
	result := make(map[string]any)
	if match.ProcessName != "" {
		result["process_name"] = match.ProcessName
	}
	if match.ProcessNameType != "" {
		result["process_name_type"] = string(match.ProcessNameType)
	}
	if match.ParentName != "" {
		result["parent_name"] = match.ParentName
	}
	if match.ParentNameType != "" {
		result["parent_name_type"] = string(match.ParentNameType)
	}
	if match.PID != 0 {
		result["pid"] = match.PID
	}
	if match.PPID != 0 {
		result["ppid"] = match.PPID
	}
	if match.Filename != "" {
		result["filename"] = match.Filename
	}
	if match.DestPort != 0 {
		result["dest_port"] = match.DestPort
	}
	if match.DestIP != "" {
		result["dest_ip"] = match.DestIP
	}
	if match.CgroupID != "" {
		result["cgroup_id"] = match.CgroupID
	}
	return result
}

func decodePolicyRule(raw map[string]any) (policy.Rule, error) {
	var rule policy.Rule
	if raw == nil {
		return rule, fmt.Errorf("missing rule payload")
	}
	if value, ok := raw["name"].(string); ok {
		rule.Name = value
	}
	if value, ok := raw["description"].(string); ok {
		rule.Description = value
	}
	if value, ok := raw["severity"].(string); ok {
		rule.Severity = value
	}
	if value, ok := raw["action"].(string); ok {
		rule.Action = policy.ActionType(value)
	}
	if value, ok := raw["type"].(string); ok {
		rule.Type = policy.RuleType(value)
	}
	if value, ok := raw["state"].(string); ok {
		rule.State = policy.RuleState(value)
	}
	if value, ok := raw["mode"].(string); ok && rule.State == "" {
		rule.State = policy.RuleState(value)
	}
	if matchRaw, ok := raw["match"].(map[string]any); ok {
		if value, ok := matchRaw["process_name"].(string); ok {
			rule.Match.ProcessName = value
		}
		if value, ok := matchRaw["process_name_type"].(string); ok {
			rule.Match.ProcessNameType = policy.MatchType(value)
		}
		if value, ok := matchRaw["parent_name"].(string); ok {
			rule.Match.ParentName = value
		}
		if value, ok := matchRaw["parent_name_type"].(string); ok {
			rule.Match.ParentNameType = policy.MatchType(value)
		}
		if value, ok := matchRaw["filename"].(string); ok {
			rule.Match.Filename = value
		}
		if value, ok := matchRaw["dest_ip"].(string); ok {
			rule.Match.DestIP = value
		}
		if value, ok := matchRaw["cgroup_id"].(string); ok {
			rule.Match.CgroupID = value
		}
		if value, ok := numberToUint32(matchRaw["pid"]); ok {
			rule.Match.PID = value
		}
		if value, ok := numberToUint32(matchRaw["ppid"]); ok {
			rule.Match.PPID = value
		}
		if value, ok := numberToUint16(matchRaw["dest_port"]); ok {
			rule.Match.DestPort = value
		}
	}
	return rule, nil
}

func numberToUint16(value any) (uint16, bool) {
	switch v := value.(type) {
	case float64:
		return uint16(v), true
	case int:
		return uint16(v), true
	case int64:
		return uint16(v), true
	default:
		return 0, false
	}
}

func numberToUint32(value any) (uint32, bool) {
	switch v := value.(type) {
	case float64:
		return uint32(v), true
	case int:
		return uint32(v), true
	case int64:
		return uint32(v), true
	default:
		return 0, false
	}
}
