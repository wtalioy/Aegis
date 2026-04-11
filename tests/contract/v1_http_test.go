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

func TestOldAPIPathsAreNotRegistered(t *testing.T) {
	runtime := newRuntime(t)
	handler := httpapi.NewHandler(httpapi.DependenciesFromRuntime(runtime), nil)

	req := httptest.NewRequest(http.MethodGet, "/api"+"/stats", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected old route to be removed, got %d", rec.Code)
	}
}
