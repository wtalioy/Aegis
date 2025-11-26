package ui

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"eulerguard/pkg/config"
	"eulerguard/pkg/events"
	"eulerguard/pkg/proc"
	"eulerguard/pkg/profiler"
	"eulerguard/pkg/rules"
	"eulerguard/pkg/tracer"
	"eulerguard/pkg/workload"
)

type App struct {
	ctx  context.Context
	opts config.Options

	processTree      *proc.ProcessTree
	ruleEngine       *rules.Engine
	workloadRegistry *workload.Registry
	core             *tracer.Core

	stats  *Stats
	bridge *Bridge

	profiler *profiler.Profiler
	learning struct {
		active    bool
		startTime time.Time
		duration  time.Duration
	}

	ready        chan struct{}
	stopWatcher  chan struct{}
	watcherMu    sync.Mutex
	lastRuleMod  time.Time
}

func NewApp(opts config.Options) *App {
	stats := NewStats()
	return &App{
		opts:        opts,
		stats:       stats,
		bridge:      NewBridge(stats),
		ready:       make(chan struct{}),
		stopWatcher: make(chan struct{}),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	close(a.ready)
}

func (a *App) WaitForReady() { <-a.ready }

func (a *App) Shutdown(ctx context.Context) {
	if a.profiler != nil {
		a.profiler.Stop()
	}
	close(a.stopWatcher)
}

func (a *App) Bridge() *Bridge { return a.bridge }
func (a *App) Stats() *Stats   { return a.stats }

func (a *App) Run() error {
	log.Println("Starting eBPF tracer...")

	core, err := tracer.Init(a.opts)
	if err != nil {
		return err
	}
	defer core.Close()

	a.core = core
	a.processTree = core.ProcessTree
	a.ruleEngine = core.RuleEngine
	a.workloadRegistry = core.WorkloadRegistry

	a.stats.SetWorkloadCountFunc(core.WorkloadRegistry.Count)

	a.bridge.SetRuleEngine(core.ProcessTree, core.RuleEngine)
	a.bridge.SetWorkloadRegistry(core.WorkloadRegistry)

	go a.watchRulesFile()

	chain := events.NewHandlerChain()
	chain.Add(a.bridge)

	log.Println("eBPF tracer started")
	return tracer.EventLoop(core.Reader, chain, core.ProcessTree, core.WorkloadRegistry)
}

func (a *App) watchRulesFile() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	if info, err := os.Stat(a.opts.RulesPath); err == nil {
		a.lastRuleMod = info.ModTime()
	}

	for {
		select {
		case <-a.stopWatcher:
			return
		case <-ticker.C:
			info, err := os.Stat(a.opts.RulesPath)
			if err != nil {
				continue
			}

			a.watcherMu.Lock()
			if info.ModTime().After(a.lastRuleMod) {
				a.lastRuleMod = info.ModTime()
				a.watcherMu.Unlock()

				if err := a.reloadRules(); err != nil {
					log.Printf("Failed to reload rules: %v", err)
				} else {
					log.Println("Rules reloaded due to file change")
					a.bridge.emit("rules:reload", nil)
				}
			} else {
				a.watcherMu.Unlock()
			}
		}
	}
}

func (a *App) reloadRules() error {
	if a.core == nil {
		return nil
	}

	if err := a.core.ReloadRules(); err != nil {
		return err
	}

	a.ruleEngine = a.core.RuleEngine
	a.bridge.SetRuleEngine(a.processTree, a.core.RuleEngine)

	return nil
}
