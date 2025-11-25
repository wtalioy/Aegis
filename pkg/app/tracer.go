package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"eulerguard/pkg/config"
	"eulerguard/pkg/ebpf"
	"eulerguard/pkg/events"
	"eulerguard/pkg/handlers"
	"eulerguard/pkg/metrics"
	"eulerguard/pkg/output"
	"eulerguard/pkg/proctree"
	"eulerguard/pkg/profiler"
	"eulerguard/pkg/rules"
	"eulerguard/pkg/utils"

	cebpf "github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

type Tracer struct {
	opts        config.Options
	handlers    *events.HandlerChain
	profiler    *profiler.Profiler
	processTree *proctree.ProcessTree
}

func NewTracer(opts config.Options) *Tracer {
	return &Tracer{
		opts:     opts,
		handlers: events.NewHandlerChain(),
	}
}

func NewExecveTracer(opts config.Options) *Tracer {
	return NewTracer(opts)
}

func (t *Tracer) Run(ctx context.Context) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root (current euid=%d)", os.Geteuid())
	}

	t.processTree = proctree.NewProcessTree(
		t.opts.ProcessTreeMaxAge,
		t.opts.ProcessTreeMaxSize,
		t.opts.ProcessTreeMaxChainLength,
	)

	if t.opts.LearnMode {
		t.profiler = profiler.NewProfiler()
		t.handlers.Add(t.profiler)
		log.Printf("Learning mode enabled for %v, output will be written to %s",
			t.opts.LearnDuration, t.opts.LearnOutputPath)
	}

	objs, err := ebpf.LoadExecveObjects(t.opts.BPFPath, t.opts.RingBufferSize)
	if err != nil {
		return err
	}
	defer objs.Close()

	links, err := t.attachTracepoints(objs)
	if err != nil {
		return err
	}
	defer closeLinks(links)

	reader, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		return fmt.Errorf("open ringbuf reader: %w", err)
	}
	defer reader.Close()

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if t.opts.LearnMode {
		go t.runLearnModeTimer(runCtx, cancel)
	}

	go func() {
		<-runCtx.Done()
		_ = reader.Close()
	}()

	meter := metrics.NewRateMeter(2 * time.Second)
	printer, err := output.NewPrinter(t.opts.JSONLines, meter, t.opts.LogFile, t.opts.LogBufferSize)
	if err != nil {
		return fmt.Errorf("failed to create printer: %w", err)
	}
	defer printer.Close()

	loadedRules, ruleEngine := t.loadRules()
	if err := populateMonitoredPaths(objs.MonitoredPaths, loadedRules, t.opts.RulesPath); err != nil {
		return fmt.Errorf("failed to populate monitored paths: %w", err)
	}

	alertHandler := handlers.NewAlertHandler(t.processTree, printer, ruleEngine)
	t.handlers.Add(alertHandler)

	t.logStartup()

	return t.eventLoop(reader)
}

func (t *Tracer) attachTracepoints(objs *ebpf.ExecveObjects) ([]link.Link, error) {
	var links []link.Link

	tp, err := link.Tracepoint("sched", "sched_process_exec", objs.HandleExec, nil)
	if err != nil {
		return nil, fmt.Errorf("attach tracepoint exec: %w", err)
	}
	links = append(links, tp)

	tpOpenat, err := link.Tracepoint("syscalls", "sys_enter_openat", objs.TracepointOpenat, nil)
	if err != nil {
		closeLinks(links)
		return nil, fmt.Errorf("attach tracepoint openat: %w", err)
	}
	links = append(links, tpOpenat)

	tpConnect, err := link.Tracepoint("syscalls", "sys_enter_connect", objs.TracepointConnect, nil)
	if err != nil {
		closeLinks(links)
		return nil, fmt.Errorf("attach tracepoint connect: %w", err)
	}
	links = append(links, tpConnect)

	return links, nil
}

// closeLinks closes all tracepoint links.
func closeLinks(links []link.Link) {
	for _, l := range links {
		_ = l.Close()
	}
}

func (t *Tracer) runLearnModeTimer(ctx context.Context, cancel context.CancelFunc) {
	timer := time.NewTimer(t.opts.LearnDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		log.Printf("Learning period complete. Collected %d unique behavior profiles.",
			t.profiler.Count())
		t.profiler.Stop()

		if err := t.profiler.SaveYAML(t.opts.LearnOutputPath); err != nil {
			log.Printf("Error saving whitelist rules: %v", err)
		} else {
			log.Printf("Whitelist rules saved to %s", t.opts.LearnOutputPath)
		}
		cancel()
	case <-ctx.Done():
		return
	}
}

func (t *Tracer) loadRules() ([]rules.Rule, *rules.Engine) {
	loadedRules, err := rules.LoadRules(t.opts.RulesPath)
	if err != nil {
		log.Printf("Warning: failed to load rules from %s: %v", t.opts.RulesPath, err)
		log.Printf("Continuing without rules...")
		loadedRules = []rules.Rule{}
	} else {
		log.Printf("Loaded %d detection rules from %s", len(loadedRules), t.opts.RulesPath)
	}
	return loadedRules, rules.NewEngine(loadedRules)
}

func (t *Tracer) logStartup() {
	if t.opts.LearnMode {
		log.Printf("EulerGuard learning mode started (BPF object: %s)", t.opts.BPFPath)
	} else {
		log.Printf("EulerGuard tracer ready (BPF object: %s, monitoring paths from %s)",
			t.opts.BPFPath, t.opts.RulesPath)
	}
}

// read events from ring buffer and dispatch them to handlers.
func (t *Tracer) eventLoop(reader *ringbuf.Reader) error {
	for {
		record, err := reader.Read()
		if errors.Is(err, ringbuf.ErrClosed) {
			return nil
		}
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			return fmt.Errorf("read ringbuf: %w", err)
		}

		if len(record.RawSample) < 1 {
			continue
		}

		t.dispatchEvent(record.RawSample)
	}
}

// decode and dispatch an event to all handlers.
func (t *Tracer) dispatchEvent(data []byte) {
	switch events.EventType(data[0]) {
	case events.EventTypeExec:
		ev, err := events.DecodeExecEvent(data)
		if err != nil {
			log.Printf("Error decoding exec event: %v", err)
			return
		}
		// Update process tree before dispatching
		t.processTree.AddProcess(ev.PID, ev.PPID, ev.CgroupID, utils.ExtractCString(ev.Comm[:]))
		t.handlers.HandleExec(ev)

	case events.EventTypeFileOpen:
		ev, err := events.DecodeFileOpenEvent(data)
		if err != nil {
			log.Printf("Error decoding file open event: %v", err)
			return
		}
		filename := utils.ExtractCString(ev.Filename[:])
		t.handlers.HandleFileOpen(ev, filename)

	case events.EventTypeConnect:
		ev, err := events.DecodeConnectEvent(data)
		if err != nil {
			log.Printf("Error decoding connect event: %v", err)
			return
		}
		t.handlers.HandleConnect(ev)
	}
}

// add file paths from rules to BPF map for filtering.
func populateMonitoredPaths(bpfMap *cebpf.Map, ruleList []rules.Rule, rulesPath string) error {
	if bpfMap == nil {
		return fmt.Errorf("monitored_paths map is nil")
	}

	pathSet := make(map[string]struct{})

	for _, rule := range ruleList {
		if rule.Match.Filename != "" {
			pathSet[rule.Match.Filename] = struct{}{}
		}
		if rule.Match.FilePath != "" {
			pathSet[rule.Match.FilePath] = struct{}{}
		}
	}

	if len(pathSet) == 0 {
		log.Printf("Warning: No file access rules found in %s", rulesPath)
		return nil
	}

	count := 0
	value := uint8(1)

	for path := range pathSet {
		key := make([]byte, events.PathMaxLen)
		copy(key, []byte(path))
		if err := bpfMap.Put(key, value); err != nil {
			return fmt.Errorf("failed to add path %q to BPF map: %w", path, err)
		}
		count++
	}

	log.Printf("Populated BPF map with %d monitored paths from %s", count, rulesPath)
	return nil
}
