package ui

import (
	"context"
	"log"
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

	stats  *Stats
	bridge *Bridge

	profiler *profiler.Profiler
	learning struct {
		active    bool
		startTime time.Time
		duration  time.Duration
	}

	ready chan struct{}
}

func NewApp(opts config.Options) *App {
	stats := NewStats()
	return &App{
		opts:   opts,
		stats:  stats,
		bridge: NewBridge(stats),
		ready:  make(chan struct{}),
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

	a.processTree = core.ProcessTree
	a.ruleEngine = core.RuleEngine
	a.workloadRegistry = core.WorkloadRegistry

	// Wire workload count function to stats
	a.stats.SetWorkloadCountFunc(core.WorkloadRegistry.Count)

	a.bridge.SetRuleEngine(core.ProcessTree, core.RuleEngine)
	a.bridge.SetWorkloadRegistry(core.WorkloadRegistry)

	chain := events.NewHandlerChain()
	chain.Add(a.bridge)

	log.Println("eBPF tracer started")
	return tracer.EventLoop(core.Reader, chain, core.ProcessTree, core.WorkloadRegistry)
}
