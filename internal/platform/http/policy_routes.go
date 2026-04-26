package httpapi

import (
	"net/http"
	"strings"
	"time"

	"aegis/internal/policy"
	"aegis/internal/policy/rules"

	"gopkg.in/yaml.v3"
)

type policyMatchDTO struct {
	ProcessName     string `json:"processName,omitempty"`
	ProcessNameType string `json:"processNameType,omitempty"`
	ParentName      string `json:"parentName,omitempty"`
	ParentNameType  string `json:"parentNameType,omitempty"`
	PID             uint32 `json:"pid,omitempty"`
	PPID            uint32 `json:"ppid,omitempty"`
	Filename        string `json:"filename,omitempty"`
	DestPort        uint16 `json:"destPort,omitempty"`
	DestIP          string `json:"destIp,omitempty"`
	CgroupID        string `json:"cgroupId,omitempty"`
}

type policyRuleDTO struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Severity    string         `json:"severity"`
	Action      string         `json:"action"`
	Type        string         `json:"type"`
	State       string         `json:"state"`
	Match       policyMatchDTO `json:"match"`
	YAML        string         `json:"yaml"`
	CreatedAt   time.Time      `json:"createdAt,omitempty"`
	DeployedAt  *time.Time     `json:"deployedAt,omitempty"`
	PromotedAt  *time.Time     `json:"promotedAt,omitempty"`
}

type policyWriteRequest struct {
	Rule policyRuleDTO `json:"rule"`
}

type policyWriteResponse struct {
	Success bool          `json:"success"`
	Rule    policyRuleDTO `json:"rule"`
}

type policySuccessResponse struct {
	Success bool `json:"success"`
}

type testingRuleDTO struct {
	policyRuleDTO
	Validation policy.PromotionReadiness `json:"validation"`
	Stats      policy.TestingStats       `json:"stats"`
}

func registerPolicyRoutes(mux *http.ServeMux, deps Dependencies) {
	registerAliases(mux, []string{"/api/v1/policies/testing"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		items := deps.Policy.TestingRules()
		payload := make([]testingRuleDTO, 0, len(items))
		for _, item := range items {
			payload = append(payload, testingRuleDTO{
				policyRuleDTO: toPolicyRuleDTO(item.Rule),
				Validation:    item.Validation,
				Stats:         item.Stats,
			})
		}
		writeJSON(w, http.StatusOK, payload)
	})

	registerAliasesWithPrefix(mux, []string{"/api/v1/policies/validation/"}, func(w http.ResponseWriter, r *http.Request, name string) {
		setCORS(w)
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		readiness, stats, err := deps.Policy.Validation(name)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		rule, ok := deps.Policy.Get(name)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, testingRuleDTO{
			policyRuleDTO: toPolicyRuleDTOValue(rule),
			Validation:    readiness,
			Stats:         stats,
		})
	})

	registerAliases(mux, []string{"/api/v1/policies"}, func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		switch r.Method {
		case http.MethodGet:
			ruleList := deps.Policy.List()
			payload := make([]policyRuleDTO, 0, len(ruleList))
			for _, rule := range ruleList {
				payload = append(payload, toPolicyRuleDTO(rule))
			}
			writeJSON(w, http.StatusOK, payload)
		case http.MethodPost:
			var req policyWriteRequest
			if err := decodeJSON(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			created, err := deps.Policy.Create(fromPolicyRuleDTO(req.Rule))
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, policyWriteResponse{Success: true, Rule: toPolicyRuleDTO(created)})
		case http.MethodOptions:
			allowJSONOptions(w, http.MethodGet, http.MethodPost)
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
			if !requireMethod(w, r, http.MethodPost) {
				return
			}
			if err := deps.Policy.Promote(name); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusOK, policySuccessResponse{Success: true})
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
			writeJSON(w, http.StatusOK, toPolicyRuleDTOValue(rule))
		case http.MethodPut:
			var req policyWriteRequest
			if err := decodeJSON(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			updated, err := deps.Policy.Update(name, fromPolicyRuleDTO(req.Rule))
			if err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusOK, policyWriteResponse{Success: true, Rule: toPolicyRuleDTO(updated)})
		case http.MethodDelete:
			if err := deps.Policy.Delete(name); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusOK, policySuccessResponse{Success: true})
		case http.MethodOptions:
			allowJSONOptions(w, http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPost)
		default:
			methodNotAllowed(w)
		}
	})
}

func toPolicyRuleDTO(rule policy.Rule) policyRuleDTO {
	cleanRule := rules.CleanRuleForYAML(rule)
	yamlBytes, _ := yaml.Marshal(cleanRule)

	return policyRuleDTO{
		Name:        rule.Name,
		Description: rule.Description,
		Severity:    rule.Severity,
		Action:      string(rule.Action),
		Type:        string(rule.DeriveType()),
		State:       string(rule.State),
		Match: policyMatchDTO{
			ProcessName:     rule.Match.ProcessName,
			ProcessNameType: string(rule.Match.ProcessNameType),
			ParentName:      rule.Match.ParentName,
			ParentNameType:  string(rule.Match.ParentNameType),
			PID:             rule.Match.PID,
			PPID:            rule.Match.PPID,
			Filename:        rule.Match.Filename,
			DestPort:        rule.Match.DestPort,
			DestIP:          rule.Match.DestIP,
			CgroupID:        rule.Match.CgroupID,
		},
		YAML:       string(yamlBytes),
		CreatedAt:  rule.CreatedAt,
		DeployedAt: rule.DeployedAt,
		PromotedAt: rule.PromotedAt,
	}
}

func toPolicyRuleDTOValue(rule *policy.Rule) policyRuleDTO {
	if rule == nil {
		return policyRuleDTO{}
	}
	return toPolicyRuleDTO(*rule)
}

func fromPolicyRuleDTO(dto policyRuleDTO) policy.Rule {
	return policy.Rule{
		Name:        dto.Name,
		Description: dto.Description,
		Severity:    dto.Severity,
		Action:      policy.ActionType(dto.Action),
		Type:        policy.RuleType(dto.Type),
		State:       policy.RuleState(dto.State),
		Match: policy.MatchCondition{
			ProcessName:     dto.Match.ProcessName,
			ProcessNameType: policy.MatchType(dto.Match.ProcessNameType),
			ParentName:      dto.Match.ParentName,
			ParentNameType:  policy.MatchType(dto.Match.ParentNameType),
			PID:             dto.Match.PID,
			PPID:            dto.Match.PPID,
			Filename:        dto.Match.Filename,
			DestPort:        dto.Match.DestPort,
			DestIP:          dto.Match.DestIP,
			CgroupID:        dto.Match.CgroupID,
		},
	}
}
