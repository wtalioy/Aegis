package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"aegis/pkg/rules"
	"aegis/pkg/server"
)

type ValidationResponse struct {
	Rule       *rules.Rule              `json:"rule"`
	Validation rules.PromotionReadiness `json:"validation"`
	Stats      rules.TestingStats       `json:"stats"`
}

type TestingRuleResponse struct {
	Rule       *rules.Rule              `json:"rule"`
	Validation rules.PromotionReadiness `json:"validation"`
	Stats      rules.TestingStats       `json:"stats"`
}

func RegisterRulesHandlers(mux *http.ServeMux, app *server.App) {
	mux.HandleFunc("/api/rules", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		}

		if r.Method == "GET" {
			rules := app.GetRules()
			data, _ := json.Marshal(rules)
			w.Write(data)
			return
		}

		if r.Method == "POST" {
			// Create rule
			var req struct {
				Rule rules.Rule `json:"rule"`
				Mode string     `json:"mode"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Set state if provided (map legacy "mode" to "state")
			if req.Mode != "" {
				switch req.Mode {
				case "draft":
					req.Rule.State = rules.RuleStateDraft
				case "testing":
					req.Rule.State = rules.RuleStateTesting
				case "production":
					req.Rule.State = rules.RuleStateProduction
				}
			} else if req.Rule.State == "" {
				// Default to testing if no state is set
				req.Rule.State = rules.RuleStateTesting
			}

			// Load existing rules and append new rule
			allRules := app.GetRulesInternal()

			// Set timestamps for new rule
			now := time.Now()
			if req.Rule.CreatedAt.IsZero() {
				req.Rule.CreatedAt = now
			}
			// Set DeployedAt if deploying to testing or production
			if req.Rule.State == rules.RuleStateTesting || req.Rule.State == rules.RuleStateProduction {
				if req.Rule.DeployedAt == nil {
					req.Rule.DeployedAt = &now
				}
			}

			allRules = append(allRules, req.Rule)

			// Save and reload
			if err := app.SaveAndReloadRules(allRules); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"rule":    req.Rule,
			})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Phase 2: Update rule
	mux.HandleFunc("/api/rules/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "PUT, DELETE, POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		}

		// Extract rule name from path
		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/rules/"), "/")
		if len(pathParts) == 0 {
			http.Error(w, "Invalid rule name", http.StatusBadRequest)
			return
		}
		ruleName := pathParts[0]

		// Update rule
		if r.Method == "PUT" {
			var req struct {
				Rule rules.Rule `json:"rule"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Get rules and update
			rule, allRules := app.FindRuleByName(ruleName)
			if rule == nil {
				http.Error(w, "Rule not found", http.StatusNotFound)
				return
			}

			// Preserve existing metadata fields
			existingCreatedAt := rule.CreatedAt
			existingDeployedAt := rule.DeployedAt
			existingPromotedAt := rule.PromotedAt
			existingLastReviewedAt := rule.LastReviewedAt
			existingReviewNotes := rule.ReviewNotes

			// Preserve the rule name (must match URL parameter)
			// Update the rule fields
			// Note: Name is preserved from URL parameter, not from request body
			rule.Description = req.Rule.Description
			rule.Action = req.Rule.Action
			rule.Severity = req.Rule.Severity
			rule.Match = req.Rule.Match
			rule.Type = req.Rule.Type
			rule.State = req.Rule.State

			// Preserve metadata if not provided
			if existingCreatedAt.IsZero() && !req.Rule.CreatedAt.IsZero() {
				rule.CreatedAt = req.Rule.CreatedAt
			} else if !existingCreatedAt.IsZero() {
				rule.CreatedAt = existingCreatedAt
			}

			if existingDeployedAt != nil {
				rule.DeployedAt = existingDeployedAt
			} else if req.Rule.DeployedAt != nil {
				rule.DeployedAt = req.Rule.DeployedAt
			}

			if existingPromotedAt != nil {
				rule.PromotedAt = existingPromotedAt
			} else if req.Rule.PromotedAt != nil {
				rule.PromotedAt = req.Rule.PromotedAt
			}

			if existingLastReviewedAt != nil {
				rule.LastReviewedAt = existingLastReviewedAt
			} else if req.Rule.LastReviewedAt != nil {
				rule.LastReviewedAt = req.Rule.LastReviewedAt
			}

			if existingReviewNotes != "" {
				rule.ReviewNotes = existingReviewNotes
			} else {
				rule.ReviewNotes = req.Rule.ReviewNotes
			}

			// Update DeployedAt if state changed to testing or production
			if (rule.State == rules.RuleStateTesting || rule.State == rules.RuleStateProduction) && rule.DeployedAt == nil {
				now := time.Now()
				rule.DeployedAt = &now
			}

			if err := app.SaveAndReloadRules(allRules); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"rule":    *rule,
			})
			return
		}

		// Delete rule
		if r.Method == "DELETE" {
			// Get rules and delete
			rule, allRules := app.FindRuleByName(ruleName)
			if rule == nil {
				http.Error(w, "Rule not found", http.StatusNotFound)
				return
			}
			// Remove rule from slice
			for i, r := range allRules {
				if r.Name == ruleName {
					allRules = append(allRules[:i], allRules[i+1:]...)
					break
				}
			}

			if err := app.SaveAndReloadRules(allRules); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// NEW: Validation endpoints
	mux.HandleFunc("/api/rules/validation/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		}

		// Extract rule name from path
		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/rules/validation/"), "/")
		if len(pathParts) == 0 {
			http.Error(w, "Invalid rule name", http.StatusBadRequest)
			return
		}

		// Handle sub-paths
		if len(pathParts) > 1 {
			subPath := pathParts[1]

			// POST /api/rules/validation/{name}/promote
			if subPath == "promote" && r.Method == "POST" {
				handlePromoteRule(w, r, app)
				return
			}

			// POST /api/rules/validation/{name}/demote
			if subPath == "demote" && r.Method == "POST" {
				handleDemoteRule(w, r, app)
				return
			}

			// Invalid sub-path
			http.Error(w, "Invalid endpoint", http.StatusBadRequest)
			return
		}

		// GET /api/rules/validation/{name}
		if r.Method == "GET" {
			handleGetRuleValidation(w, r, app)
			return
		}

		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
	})

	// GET /api/rules/testing
	mux.HandleFunc("/api/rules/testing", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		}

		if r.Method == "GET" {
			handleGetTestingRules(w, r, app)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}


func handleGetRuleValidation(w http.ResponseWriter, r *http.Request, app *server.App) {
	setCORS(w)
	// Extract rule ID from URL path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/rules/validation/"), "/")
	if len(pathParts) == 0 {
		http.Error(w, "Invalid rule name", http.StatusBadRequest)
		return
	}
	ruleID := pathParts[0]
	// URL decode the rule name
	if decoded, err := url.QueryUnescape(ruleID); err == nil {
		ruleID = decoded
	}

	// Get rule from engine
	rule, _ := app.FindRuleByName(ruleID)
	if rule == nil {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
	}

	// Get validation status
	testingBuffer := app.GetTestingBuffer()
	if testingBuffer == nil {
		http.Error(w, "testing buffer not available", http.StatusInternalServerError)
		return
	}
	opts := app.Options()
	validationService := rules.NewValidationService(testingBuffer, opts.PromotionMinObservationMinutes, opts.PromotionMinHits)
	validation := validationService.CalculatePromotionReadiness(rule)

	// Get testing stats
	testingStats := testingBuffer.GetStats(rule.Name)

	response := ValidationResponse{
		Rule:       rule,
		Validation: validation,
		Stats:      testingStats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func handlePromoteRule(w http.ResponseWriter, r *http.Request, app *server.App) {
	setCORS(w)
	// Extract rule ID from URL path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/rules/validation/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "Invalid rule name", http.StatusBadRequest)
		return
	}
	ruleID := pathParts[0]

	// Get rule from engine
	rule, _ := app.FindRuleByName(ruleID)
	if rule == nil {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
	}

	if !rule.IsTesting() {
		http.Error(w, "Rule must be in testing mode to promote", http.StatusBadRequest)
		return
	}

	// Parse request body for force flag
	var reqBody struct {
		Force bool `json:"force"`
	}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&reqBody)
	}

	// Check readiness (unless force is true)
	if !reqBody.Force {
		testingBuffer := app.GetTestingBuffer()
		if testingBuffer == nil {
			http.Error(w, "testing buffer not available", http.StatusInternalServerError)
			return
		}
		opts := app.Options()
		validationService := rules.NewValidationService(testingBuffer, opts.PromotionMinObservationMinutes, opts.PromotionMinHits)
		readiness := validationService.CalculatePromotionReadiness(rule)
		if !readiness.IsReady {
			http.Error(w, fmt.Sprintf("Rule is not ready for promotion. Score: %.0f%%. Use force=true to promote anyway.", readiness.Score*100), http.StatusBadRequest)
			return
		}
	}

	// Promote rule using existing API (updates YAML and reloads)
	// PromoteRule will set State, Mode, and PromotedAt
	if err := app.PromoteRule(ruleID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true, "promoted": ruleID})
}


func handleGetTestingRules(w http.ResponseWriter, _ *http.Request, app *server.App) {
	setCORS(w)
	allRules := app.GetRulesInternal()

	// Filter to testing rules only
	var testingRules []*rules.Rule
	for i := range allRules {
		// Check both IsTesting() and explicit state comparison
		if allRules[i].IsTesting() || allRules[i].State == rules.RuleStateTesting {
			testingRules = append(testingRules, &allRules[i])
		}
	}

	// Enrich with validation data
	response := make([]TestingRuleResponse, len(testingRules))
	testingBuffer := app.GetTestingBuffer()
	if testingBuffer == nil {
		http.Error(w, "testing buffer not available", http.StatusInternalServerError)
		return
	}
	opts := app.Options()
	validationService := rules.NewValidationService(testingBuffer, opts.PromotionMinObservationMinutes, opts.PromotionMinHits)

	for i, rule := range testingRules {
		validation := validationService.CalculatePromotionReadiness(rule)
		stats := testingBuffer.GetStats(rule.Name)

		response[i] = TestingRuleResponse{
			Rule:       rule,
			Validation: validation,
			Stats:      stats,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func handleDemoteRule(w http.ResponseWriter, r *http.Request, app *server.App) {
	setCORS(w)
	// Extract rule ID from URL path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/rules/validation/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "Invalid rule name", http.StatusBadRequest)
		return
	}
	ruleID := pathParts[0]

	// Get rule from engine
	// Find rule by name in current rules
	core := app.Core()
	if core == nil || core.RuleEngine == nil {
		http.Error(w, "rule engine not available", http.StatusInternalServerError)
		return
	}
	var rule *rules.Rule
	allRules := core.RuleEngine.GetRules()
	for i := range allRules {
		if allRules[i].Name == ruleID {
			rule = &allRules[i]
			break
		}
	}
	if rule == nil {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
	}

	if !rule.IsProduction() {
		http.Error(w, "Rule must be in production mode to demote", http.StatusBadRequest)
		return
	}

	// Demote rule
	now := time.Now()
	rule.State = rules.RuleStateTesting
	rule.DeployedAt = &now
	rule.ActualTestingHits = 0 // Reset testing hit count

	// Persist changes and reload
	if err := app.SaveAndReloadRules(allRules); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Clear testing hits for this rule
	if testingBuffer := app.GetTestingBuffer(); testingBuffer != nil {
		testingBuffer.ClearHits(ruleID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true, "demoted": ruleID})
}
