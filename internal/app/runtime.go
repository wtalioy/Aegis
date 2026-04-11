package app

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"syscall"
	"time"

	internalanalysis "aegis/internal/analysis"
	internalconfig "aegis/internal/platform/config"
	internalebpf "aegis/internal/platform/ebpf"
	"aegis/internal/platform/persistence"
	"aegis/internal/policy"
	"aegis/internal/shared/stream"
	"aegis/internal/system"
	"aegis/internal/telemetry"
	"aegis/internal/telemetry/proc"
	"aegis/internal/telemetry/workload"
	"github.com/cilium/ebpf/ringbuf"
)

type Runtime struct {
	cfg         internalconfig.Config
	configPath  string
	settings    *system.SettingsService
	stats       *system.Stats
	telemetry   *telemetry.Service
	policy      *policy.Service
	analysis    *internalanalysis.Service
	pipeline    *IngestPipeline
	eventStream *stream.Hub[telemetry.Event]
	alertStream *stream.Hub[system.Alert]
	probeState  system.ProbeStatus

	mu          sync.RWMutex
	resources   *internalebpf.Resources
	stopWatcher chan struct{}
	watcherMu   sync.Mutex
	lastRuleMod time.Time
	wg          sync.WaitGroup
}

func NewRuntime(cfg internalconfig.Config, configPath string) *Runtime {
	processTree := proc.NewProcessTree(cfg.ProcessTreeMaxAge(), cfg.Telemetry.ProcessTreeMaxSize, cfg.Telemetry.ProcessTreeMaxChainLength)
	workloads := workload.NewRegistry(1000)
	profiles := proc.NewProfileRegistry()

	stats := system.NewStats(internalconfig.DefaultMaxAlerts, internalconfig.DefaultAlertDedupWindow)
	stats.SetWorkloadCountFunc(workloads.Count)

	telemetryService := telemetry.NewService(cfg.Telemetry.RecentEventsCapacity, cfg.Telemetry.EventIndexSize, processTree, workloads, profiles)

	ruleRepo := persistence.NewRuleRepository(cfg.Policy.RulesPath)
	policyService := policy.NewService(ruleRepo, nil, cfg.Policy.PromotionMinObservationMinutes, cfg.Policy.PromotionMinHits)
	if err := policyService.Load(); err != nil {
		log.Printf("Warning: failed to load rules from %s: %v", cfg.Policy.RulesPath, err)
		if err := policyService.Bootstrap([]policy.Rule{}); err != nil {
			log.Printf("Warning: failed to bootstrap empty rule set: %v", err)
		}
	}

	analysisService, err := internalanalysis.NewService(cfg, telemetryService, policyService, stats)
	if err != nil {
		log.Printf("[AI] Failed to initialize: %v", err)
	}

	runtime := &Runtime{
		cfg:         cfg,
		configPath:  configPath,
		settings:    system.NewSettingsService(cfg, persistence.NewConfigRepository(configPath)),
		stats:       stats,
		telemetry:   telemetryService,
		policy:      policyService,
		analysis:    analysisService,
		eventStream: stream.NewHub[telemetry.Event](),
		alertStream: stream.NewHub[system.Alert](),
		probeState:  system.ProbeStatus{Status: system.ProbeStatusStarting},
		stopWatcher: make(chan struct{}),
	}
	runtime.settings.SetApplier(runtime)
	return runtime
}

func (r *Runtime) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.resources != nil {
		r.mu.Unlock()
		return nil
	}
	r.mu.Unlock()
	r.setProbeStatus(system.ProbeStatusStarting, "")

	resources, err := internalebpf.Load(r.cfg)
	if err != nil {
		r.setProbeStatus(system.ProbeStatusError, err.Error())
		return err
	}

	r.mu.Lock()
	r.resources = resources
	r.mu.Unlock()

	r.policy.SetKernelSync(internalebpf.NewKernelSync(resources, r.cfg.Policy.RulesPath))
	if err := r.policy.Reload(); err != nil {
		log.Printf("Warning: failed to sync policy rules into kernel maps: %v", err)
	}

	if r.analysis != nil {
		r.analysis.StartSentinel(r.cfg)
	}

	r.wg.Add(2)
	go r.runEventLoop()
	go r.watchRulesFile()
	r.setProbeStatus(system.ProbeStatusActive, "")

	go func() {
		<-ctx.Done()
		_ = r.Stop(context.Background())
	}()

	return nil
}

func (r *Runtime) Stop(ctx context.Context) error {
	closeOnce := false
	r.watcherMu.Lock()
	select {
	case <-r.stopWatcher:
	default:
		close(r.stopWatcher)
		closeOnce = true
	}
	r.watcherMu.Unlock()

	r.mu.Lock()
	resources := r.resources
	r.resources = nil
	r.probeState = system.ProbeStatus{Status: system.ProbeStatusStopped}
	r.mu.Unlock()

	if r.analysis != nil {
		r.analysis.StopSentinel()
	}
	if resources != nil {
		_ = resources.Close()
	}
	if closeOnce {
		done := make(chan struct{})
		go func() {
			defer close(done)
			r.wg.Wait()
		}()
		select {
		case <-done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (r *Runtime) Config() internalconfig.Config {
	return r.settings.Get()
}

func (r *Runtime) Settings() *system.SettingsService {
	return r.settings
}

func (r *Runtime) Stats() *system.Stats {
	return r.stats
}

func (r *Runtime) Telemetry() *telemetry.Service {
	return r.telemetry
}

func (r *Runtime) Policy() *policy.Service {
	return r.policy
}

func (r *Runtime) Analysis() *internalanalysis.Service {
	return r.analysis
}

func (r *Runtime) EventStream() *stream.Hub[telemetry.Event] {
	return r.eventStream
}

func (r *Runtime) IngestPipeline() *IngestPipeline {
	if r.pipeline == nil {
		r.pipeline = NewIngestPipeline(r.telemetry, r.policy, r.stats, r.eventStream, r.alertStream)
	}
	return r.pipeline
}

func (r *Runtime) AlertStream() *stream.Hub[system.Alert] {
	return r.alertStream
}

func (r *Runtime) ProbeStatus() system.ProbeStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.probeState
}

func (r *Runtime) ApplyConfig(oldCfg, newCfg internalconfig.Config, hotReloadedFields []string) error {
	appliedCfg := r.liveConfigAfterHotReload(oldCfg, newCfg)

	if oldCfg.Policy.PromotionMinObservationMinutes != newCfg.Policy.PromotionMinObservationMinutes ||
		oldCfg.Policy.PromotionMinHits != newCfg.Policy.PromotionMinHits {
		r.policy.UpdateThresholds(newCfg.Policy.PromotionMinObservationMinutes, newCfg.Policy.PromotionMinHits)
	}

	if r.analysisConfigChanged(oldCfg, newCfg) || r.sentinelConfigChanged(oldCfg, newCfg) {
		if err := r.reloadAnalysis(appliedCfg); err != nil {
			return err
		}
	}

	r.mu.Lock()
	r.cfg = appliedCfg
	r.mu.Unlock()
	return nil
}

func (r *Runtime) runEventLoop() {
	defer r.wg.Done()

	r.mu.RLock()
	reader := r.resources.Reader
	r.mu.RUnlock()
	if reader == nil {
		return
	}

	for {
		record, err := reader.Read()
		if errors.Is(err, ringbuf.ErrClosed) {
			return
		}
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			r.setProbeStatus(system.ProbeStatusError, err.Error())
			log.Printf("read ring buffer: %v", err)
			return
		}
		if len(record.RawSample) == 0 {
			continue
		}

		_, _, err = r.IngestPipeline().ProcessRawSample(record.RawSample)
		if err != nil {
			log.Printf("decode raw event: %v", err)
			continue
		}
	}
}

func (r *Runtime) setProbeStatus(status string, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.probeState = system.ProbeStatus{
		Status: status,
		Error:  errMsg,
	}
}

func (r *Runtime) liveConfigAfterHotReload(oldCfg, newCfg internalconfig.Config) internalconfig.Config {
	applied := oldCfg
	applied.Policy.PromotionMinObservationMinutes = newCfg.Policy.PromotionMinObservationMinutes
	applied.Policy.PromotionMinHits = newCfg.Policy.PromotionMinHits
	applied.Analysis = newCfg.Analysis
	applied.Sentinel = newCfg.Sentinel
	return applied
}

func (r *Runtime) analysisConfigChanged(oldCfg, newCfg internalconfig.Config) bool {
	return oldCfg.Analysis != newCfg.Analysis
}

func (r *Runtime) sentinelConfigChanged(oldCfg, newCfg internalconfig.Config) bool {
	return oldCfg.Sentinel != newCfg.Sentinel
}

func (r *Runtime) reloadAnalysis(cfg internalconfig.Config) error {
	var newService *internalanalysis.Service
	var err error
	if cfg.Analysis.Mode != "" && cfg.Analysis.Mode != "disabled" {
		newService, err = internalanalysis.NewService(cfg, r.telemetry, r.policy, r.stats)
		if err != nil {
			return err
		}
		if newService != nil {
			newService.StartSentinel(cfg)
		}
	}

	r.mu.Lock()
	oldService := r.analysis
	r.analysis = newService
	r.mu.Unlock()

	if oldService != nil {
		oldService.StopSentinel()
	}
	return nil
}

func (r *Runtime) watchRulesFile() {
	defer r.wg.Done()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	if info, err := os.Stat(r.cfg.Policy.RulesPath); err == nil {
		r.lastRuleMod = info.ModTime()
	}

	for {
		select {
		case <-r.stopWatcher:
			return
		case <-ticker.C:
			info, err := os.Stat(r.cfg.Policy.RulesPath)
			if err != nil {
				continue
			}

			r.watcherMu.Lock()
			if info.ModTime().After(r.lastRuleMod) {
				r.lastRuleMod = info.ModTime()
				r.watcherMu.Unlock()
				if err := r.policy.Reload(); err != nil {
					log.Printf("reload rules: %v", err)
				}
			} else {
				r.watcherMu.Unlock()
			}
		}
	}
}
