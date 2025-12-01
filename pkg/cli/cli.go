package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"eulerguard/pkg/config"
	"eulerguard/pkg/events"
	"eulerguard/pkg/metrics"
	"eulerguard/pkg/output"
	"eulerguard/pkg/profiler"
	"eulerguard/pkg/rules"
	"eulerguard/pkg/tracer"
)

type CLI struct {
	Opts     config.Options
	Handlers *events.HandlerChain
	Profiler *profiler.Profiler
	Core     *tracer.Core
}

func RunCLI(opts config.Options, ctx context.Context) error {
	cli := &CLI{
		Opts:     opts,
		Handlers: events.NewHandlerChain(),
	}

	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root (current euid=%d)", os.Geteuid())
	}

	core, err := tracer.Init(cli.Opts)
	if err != nil {
		return err
	}
	defer core.Close()
	cli.Core = core

	if cli.Opts.LearnMode {
		cli.Profiler = profiler.NewProfiler()
		cli.Handlers.Add(cli.Profiler)
		log.Printf("Learning mode enabled for %v, rules will be merged into: %s", cli.Opts.LearnDuration, cli.Opts.RulesPath)
	}

	meter := metrics.NewRateMeter(2 * time.Second)
	printer, err := output.NewPrinter(cli.Opts.JSONLines, meter, cli.Opts.LogFile, cli.Opts.LogBufferSize)
	if err != nil {
		return fmt.Errorf("create printer: %w", err)
	}
	defer printer.Close()

	cli.Handlers.Add(&cliAlertHandler{
		processTree: core.ProcessTree,
		printer:     printer,
		ruleEngine:  core.RuleEngine,
	})

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if cli.Opts.LearnMode {
		go cli.runLearnModeTimer(runCtx, cancel)
	}

	go func() {
		<-runCtx.Done()
		_ = cli.Core.Reader.Close()
	}()

	if cli.Opts.LearnMode {
		log.Printf("EulerGuard learning mode started (BPF: %s)", cli.Opts.BPFPath)
	} else {
		log.Printf("EulerGuard tracer ready (BPF: %s, rules: %s)", cli.Opts.BPFPath, cli.Opts.RulesPath)
	}

	return tracer.EventLoop(cli.Core.Reader, cli.Handlers, cli.Core.ProcessTree, cli.Core.WorkloadRegistry)
}

func (cli *CLI) runLearnModeTimer(ctx context.Context, cancel context.CancelFunc) {
	timer := time.NewTimer(cli.Opts.LearnDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		log.Printf("Learning complete. Collected %d behavior profiles.", cli.Profiler.Count())
		cli.Profiler.Stop()

		generatedRules := cli.Profiler.GenerateRules()
		if len(generatedRules) == 0 {
			log.Printf("No rules generated from learning mode")
		} else {
			existingRules, err := rules.LoadRules(cli.Opts.RulesPath)
			if err != nil {
				existingRules = []rules.Rule{}
			}

			mergedRules := rules.MergeRules(existingRules, generatedRules)
			if err := rules.SaveRules(cli.Opts.RulesPath, mergedRules); err != nil {
				log.Printf("Error saving rules: %v", err)
			} else {
				log.Printf("Merged %d new rules into %s (total: %d rules)", len(generatedRules), cli.Opts.RulesPath, len(mergedRules))
			}
		}
		cancel()
	case <-ctx.Done():
	}
}
