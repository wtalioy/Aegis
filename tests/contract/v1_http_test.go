package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"aegis/internal/app"
	internalconfig "aegis/internal/platform/config"
	httpapi "aegis/internal/platform/http"
	"aegis/internal/policy"
	"aegis/internal/system"
)

func newRuntime(t *testing.T) *app.Runtime {
	t.Helper()
	cfg := internalconfig.Default(t.TempDir())
	cfg.Analysis.Mode = "disabled"
	cfg.Policy.RulesPath = filepath.Join(t.TempDir(), "rules.yaml")
	return app.NewRuntime(cfg, filepath.Join(t.TempDir(), "config.yaml"))
}

func TestV1SystemEventsAndPoliciesContracts(t *testing.T) {
	runtime := newRuntime(t)
	handler := httpapi.NewHandler(httpapi.DependenciesFromRuntime(runtime), nil)

	testCases := []struct {
		name string
		path string
	}{
		{name: "system stats", path: "/api/v1/system/stats"},
		{name: "system settings", path: "/api/v1/system/settings"},
		{name: "system alerts", path: "/api/v1/system/alerts"},
		{name: "events", path: "/api/v1/events"},
		{name: "policies", path: "/api/v1/policies"},
		{name: "analysis status", path: "/api/v1/analysis/status"},
		{name: "sentinel insights", path: "/api/v1/sentinel/insights"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected %s to return 200, got %d with body %s", tc.path, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestV1SettingsPutReturnsRestartMetadata(t *testing.T) {
	runtime := newRuntime(t)
	handler := httpapi.NewHandler(httpapi.DependenciesFromRuntime(runtime), nil)

	cfg := runtime.Settings().Get()
	cfg.Server.Port = 4001
	body, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/system/settings", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Updated               bool     `json:"updated"`
		HotReloadedFields     []string `json:"hot_reloaded_fields"`
		RestartRequired       bool     `json:"restart_required"`
		RestartRequiredFields []string `json:"restart_required_fields"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Updated || !resp.RestartRequired {
		t.Fatalf("expected update to require restart, got %+v", resp)
	}
	if len(resp.RestartRequiredFields) == 0 || resp.RestartRequiredFields[0] != "server.port" {
		t.Fatalf("expected restart_required_fields to include server.port, got %+v", resp.RestartRequiredFields)
	}
	if len(resp.HotReloadedFields) != 0 {
		t.Fatalf("expected no hot-reloaded fields for server port change, got %+v", resp.HotReloadedFields)
	}
}

func TestV1SettingsPutReturnsHotReloadMetadata(t *testing.T) {
	runtime := newRuntime(t)
	handler := httpapi.NewHandler(httpapi.DependenciesFromRuntime(runtime), nil)

	cfg := runtime.Settings().Get()
	cfg.Policy.PromotionMinHits = 1
	body, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/system/settings", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Updated               bool     `json:"updated"`
		HotReloadedFields     []string `json:"hot_reloaded_fields"`
		RestartRequired       bool     `json:"restart_required"`
		RestartRequiredFields []string `json:"restart_required_fields"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Updated {
		t.Fatalf("expected update metadata, got %+v", resp)
	}
	if resp.RestartRequired {
		t.Fatalf("expected hot-reload-only update, got %+v", resp)
	}
	if len(resp.RestartRequiredFields) != 0 {
		t.Fatalf("expected no restart-required fields, got %+v", resp.RestartRequiredFields)
	}
	if len(resp.HotReloadedFields) != 1 || resp.HotReloadedFields[0] != "policy.promotion_min_hits" {
		t.Fatalf("expected hot_reloaded_fields to include policy.promotion_min_hits, got %+v", resp.HotReloadedFields)
	}
}

func TestV1SettingsPutRejectsInvalidConfig(t *testing.T) {
	runtime := newRuntime(t)
	handler := httpapi.NewHandler(httpapi.DependenciesFromRuntime(runtime), nil)

	cfg := runtime.Settings().Get()
	cfg.Server.Port = 0
	body, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/system/settings", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d with body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "server.port must be greater than 0") {
		t.Fatalf("expected validation error in response, got %s", rec.Body.String())
	}
}

func TestV1AlertStreamPublishesSSE(t *testing.T) {
	runtime := newRuntime(t)
	handler := httpapi.NewHandler(httpapi.DependenciesFromRuntime(runtime), nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/alerts/stream", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler.ServeHTTP(rec, req)
	}()

	time.Sleep(20 * time.Millisecond)
	runtime.AlertStream().Publish(system.Alert{
		ID:          "alert-1",
		RuleName:    "rule-1",
		ProcessName: "bash",
		CgroupID:    "77",
		Action:      "alert",
	})
	time.Sleep(20 * time.Millisecond)
	cancel()
	wg.Wait()

	body := rec.Body.String()
	if !strings.Contains(body, "\"id\":\"alert-1\"") {
		t.Fatalf("expected alert SSE payload, got %q", body)
	}
}

func TestV1SystemStatsReflectProbeLifecycle(t *testing.T) {
	runtime := newRuntime(t)
	handler := httpapi.NewHandler(httpapi.DependenciesFromRuntime(runtime), nil)

	readProbeStatus := func(t *testing.T) string {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/stats", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected stats status 200, got %d with body %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			ProbeStatus string `json:"probeStatus"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode stats response: %v", err)
		}
		return resp.ProbeStatus
	}

	if got := readProbeStatus(t); got != system.ProbeStatusStarting {
		t.Fatalf("expected initial probe status %q, got %q", system.ProbeStatusStarting, got)
	}

	if err := runtime.Stop(context.Background()); err != nil {
		t.Fatalf("stop runtime: %v", err)
	}

	if got := readProbeStatus(t); got != system.ProbeStatusStopped {
		t.Fatalf("expected stopped probe status %q, got %q", system.ProbeStatusStopped, got)
	}
}

func TestV1PolicyTestingEndpointsReturnEmptyAndPopulatedStates(t *testing.T) {
	runtime := newRuntime(t)
	handler := httpapi.NewHandler(httpapi.DependenciesFromRuntime(runtime), nil)

	t.Run("empty testing list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/policies/testing", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d with body %s", rec.Code, rec.Body.String())
		}

		var payload []map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode testing list: %v", err)
		}
		if len(payload) != 0 {
			t.Fatalf("expected no testing rules, got %+v", payload)
		}
	})

	if err := runtime.Policy().Bootstrap([]policy.Rule{
		{
			Name:        "watch-file",
			Description: "watch-file",
			Severity:    "warning",
			Action:      policy.ActionAlert,
			Type:        policy.RuleTypeFile,
			State:       policy.RuleStateTesting,
			Match: policy.MatchCondition{
				Filename: "/tmp/watch",
			},
		},
	}); err != nil {
		t.Fatalf("bootstrap rules: %v", err)
	}

	t.Run("populated testing list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/policies/testing", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d with body %s", rec.Code, rec.Body.String())
		}

		var payload []struct {
			Name       string `json:"name"`
			State      string `json:"state"`
			Validation struct {
				IsReady bool `json:"is_ready"`
			} `json:"validation"`
			Stats struct {
				Hits int `json:"hits"`
			} `json:"stats"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode testing list: %v", err)
		}
		if len(payload) != 1 || payload[0].Name != "watch-file" || payload[0].State != string(policy.RuleStateTesting) {
			t.Fatalf("unexpected testing list payload: %+v", payload)
		}
		if payload[0].Validation.IsReady || payload[0].Stats.Hits != 0 {
			t.Fatalf("expected unready testing rule with zero hits, got %+v", payload[0])
		}
	})

	t.Run("validation detail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/policies/validation/watch-file", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d with body %s", rec.Code, rec.Body.String())
		}

		var payload struct {
			Name       string `json:"name"`
			State      string `json:"state"`
			Validation struct {
				IsReady bool `json:"is_ready"`
			} `json:"validation"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode validation payload: %v", err)
		}
		if payload.Name != "watch-file" || payload.State != string(policy.RuleStateTesting) || payload.Validation.IsReady {
			t.Fatalf("unexpected validation payload: %+v", payload)
		}
	})
}

func TestOldAPIPathsAreNotRegistered(t *testing.T) {
	runtime := newRuntime(t)
	handler := httpapi.NewHandler(httpapi.DependenciesFromRuntime(runtime), nil)

	paths := []string{
		"/api/stats",
		"/api/events",
		"/api/policies",
		"/api/settings",
	}

	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected old route %s to be removed, got %d", path, rec.Code)
		}
	}
}
