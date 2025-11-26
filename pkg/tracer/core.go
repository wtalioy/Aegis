package tracer

import (
	"errors"
	"fmt"
	"log"
	"syscall"

	"eulerguard/pkg/config"
	"eulerguard/pkg/ebpf"
	"eulerguard/pkg/events"
	"eulerguard/pkg/proc"
	"eulerguard/pkg/rules"
	"eulerguard/pkg/utils"
	"eulerguard/pkg/workload"

	cebpf "github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

type Core struct {
	Objs             *ebpf.ExecveObjects
	Links            []link.Link
	Reader           *ringbuf.Reader
	Rules            []rules.Rule
	RuleEngine       *rules.Engine
	ProcessTree      *proc.ProcessTree
	WorkloadRegistry *workload.Registry
}

func Init(opts config.Options) (*Core, error) {
	c := &Core{}

	c.ProcessTree = proc.NewProcessTree(
		opts.ProcessTreeMaxAge,
		opts.ProcessTreeMaxSize,
		opts.ProcessTreeMaxChainLength,
	)

	c.WorkloadRegistry = workload.NewRegistry(1000)

	objs, err := ebpf.LoadExecveObjects(opts.BPFPath, opts.RingBufferSize)
	if err != nil {
		return nil, fmt.Errorf("load eBPF objects: %w", err)
	}
	c.Objs = objs

	links, err := AttachTracepoints(objs)
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("attach tracepoints: %w", err)
	}
	c.Links = links

	reader, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		CloseLinks(links)
		objs.Close()
		return nil, fmt.Errorf("create ringbuf reader: %w", err)
	}
	c.Reader = reader

	c.Rules, c.RuleEngine = LoadRules(opts.RulesPath)

	if err := PopulateMonitoredPaths(objs.MonitoredPaths, c.Rules, opts.RulesPath); err != nil {
		log.Printf("Warning: failed to populate monitored paths: %v", err)
	}

	return c, nil
}

func (c *Core) Close() {
	if c.Reader != nil {
		c.Reader.Close()
	}
	CloseLinks(c.Links)
	if c.Objs != nil {
		c.Objs.Close()
	}
}

func AttachTracepoints(objs *ebpf.ExecveObjects) ([]link.Link, error) {
	var links []link.Link

	tp, err := link.Tracepoint("sched", "sched_process_exec", objs.HandleExec, nil)
	if err != nil {
		return nil, fmt.Errorf("attach exec tracepoint: %w", err)
	}
	links = append(links, tp)

	tpOpenat, err := link.Tracepoint("syscalls", "sys_enter_openat", objs.TracepointOpenat, nil)
	if err != nil {
		CloseLinks(links)
		return nil, fmt.Errorf("attach openat tracepoint: %w", err)
	}
	links = append(links, tpOpenat)

	tpConnect, err := link.Tracepoint("syscalls", "sys_enter_connect", objs.TracepointConnect, nil)
	if err != nil {
		CloseLinks(links)
		return nil, fmt.Errorf("attach connect tracepoint: %w", err)
	}
	links = append(links, tpConnect)

	return links, nil
}

func CloseLinks(links []link.Link) {
	for _, l := range links {
		_ = l.Close()
	}
}

func LoadRules(rulesPath string) ([]rules.Rule, *rules.Engine) {
	loadedRules, err := rules.LoadRules(rulesPath)
	if err != nil {
		log.Printf("Warning: failed to load rules from %s: %v", rulesPath, err)
		loadedRules = []rules.Rule{}
	} else {
		log.Printf("Loaded %d detection rules from %s", len(loadedRules), rulesPath)
	}
	return loadedRules, rules.NewEngine(loadedRules)
}

func PopulateMonitoredPaths(bpfMap *cebpf.Map, ruleList []rules.Rule, rulesPath string) error {
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
			return fmt.Errorf("add path %q to BPF map: %w", path, err)
		}
		count++
	}

	log.Printf("Populated BPF map with %d monitored paths", count)
	return nil
}

func EventLoop(reader *ringbuf.Reader, handlers *events.HandlerChain, processTree *proc.ProcessTree, registry *workload.Registry) error {
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

		DispatchEvent(record.RawSample, handlers, processTree, registry)
	}
}

func DispatchEvent(data []byte, handlers *events.HandlerChain, processTree *proc.ProcessTree, registry *workload.Registry) {
	switch events.EventType(data[0]) {
	case events.EventTypeExec:
		ev, err := events.DecodeExecEvent(data)
		if err != nil {
			log.Printf("Error decoding exec event: %v", err)
			return
		}
		processTree.AddProcess(ev.PID, ev.PPID, ev.CgroupID, utils.ExtractCString(ev.Comm[:]))
		if registry != nil {
			cgroupPath := proc.ResolveCgroupPath(ev.PID, ev.CgroupID)
			registry.RecordExec(ev.CgroupID, cgroupPath)
		}
		handlers.HandleExec(ev)

	case events.EventTypeFileOpen:
		ev, err := events.DecodeFileOpenEvent(data)
		if err != nil {
			log.Printf("Error decoding file open event: %v", err)
			return
		}
		if registry != nil {
			cgroupPath := proc.ResolveCgroupPath(ev.PID, ev.CgroupID)
			registry.RecordFile(ev.CgroupID, cgroupPath)
		}
		filename := utils.ExtractCString(ev.Filename[:])
		handlers.HandleFileOpen(ev, filename)

	case events.EventTypeConnect:
		ev, err := events.DecodeConnectEvent(data)
		if err != nil {
			log.Printf("Error decoding connect event: %v", err)
			return
		}
		if registry != nil {
			cgroupPath := proc.ResolveCgroupPath(ev.PID, ev.CgroupID)
			registry.RecordConnect(ev.CgroupID, cgroupPath)
		}
		handlers.HandleConnect(ev)
	}
}
