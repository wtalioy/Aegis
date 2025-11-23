package app

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"eulerguard/pkg/config"
	"eulerguard/pkg/ebpf"
	"eulerguard/pkg/events"
	"eulerguard/pkg/metrics"
	"eulerguard/pkg/output"
	"eulerguard/pkg/proctree"
	"eulerguard/pkg/rules"
	"eulerguard/pkg/utils"

	cebpf "github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

type ExecveTracer struct {
	opts config.Options
}

func NewExecveTracer(opts config.Options) *ExecveTracer {
	return &ExecveTracer{opts: opts}
}

func (t *ExecveTracer) Run(ctx context.Context) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root (current euid=%d)", os.Geteuid())
	}

	objs, err := ebpf.LoadExecveObjects(t.opts.BPFPath, t.opts.RingBufferSize)
	if err != nil {
		return err
	}
	defer objs.Close()

	tp, err := link.Tracepoint("sched", "sched_process_exec", objs.HandleExec, nil)
	if err != nil {
		return fmt.Errorf("attach tracepoint: %w", err)
	}
	defer tp.Close()

	// Attach tracepoint for openat
	tpOpenat, err := link.Tracepoint("syscalls", "sys_enter_openat", objs.TracepointOpenat, nil)
	if err != nil {
		return fmt.Errorf("attach tracepoint openat: %w", err)
	}
	defer tpOpenat.Close()

	reader, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		return fmt.Errorf("open ringbuf reader: %w", err)
	}
	defer reader.Close()

	go func() {
		<-ctx.Done()
		_ = reader.Close()
	}()

	meter := metrics.NewRateMeter(2 * time.Second)

	printer, err := output.NewPrinter(t.opts.JSONLines, meter, t.opts.LogFile, t.opts.LogBufferSize)
	if err != nil {
		return fmt.Errorf("failed to create printer: %w", err)
	}
	defer printer.Close()

	loadedRules, err := rules.LoadRules(t.opts.RulesPath)
	if err != nil {
		log.Printf("Warning: failed to load rules from %s: %v", t.opts.RulesPath, err)
		log.Printf("Continuing without rules...")
		loadedRules = []rules.Rule{}
	} else {
		log.Printf("Loaded %d detection rules from %s", len(loadedRules), t.opts.RulesPath)
	}
	ruleEngine := rules.NewEngine(loadedRules)

	processTree := proctree.NewProcessTree(t.opts.ProcessTreeMaxAge, t.opts.ProcessTreeMaxSize, t.opts.ProcessTreeMaxChainLength)

	if err := populateMonitoredPaths(objs.MonitoredPaths, loadedRules, t.opts.RulesPath); err != nil {
		return fmt.Errorf("failed to populate monitored paths: %w", err)
	}

	log.Printf("EulerGuard tracer ready (BPF object: %s, monitoring paths from %s)", t.opts.BPFPath, t.opts.RulesPath)

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

		switch events.EventType(record.RawSample[0]) {
		case events.EventTypeExec:
			t.handleExecEvent(record.RawSample, processTree, printer, ruleEngine)
		case events.EventTypeFileOpen:
			t.handleFileOpenEvent(record.RawSample, processTree, printer, ruleEngine)
		}
	}
}

func (t *ExecveTracer) handleExecEvent(data []byte, processTree *proctree.ProcessTree,
	printer *output.Printer, ruleEngine *rules.Engine) {

	ev, err := decodeExecEvent(data)
	if err != nil {
		log.Printf("Error decoding exec event: %v", err)
		return
	}

	processTree.AddProcess(ev.PID, ev.PPID, ev.CgroupID, utils.ExtractCString(ev.Comm[:]))
	processedEvent := printer.Print(ev)
	for _, alert := range ruleEngine.Match(processedEvent) {
		printer.PrintAlert(alert)
	}
}

func (t *ExecveTracer) handleFileOpenEvent(data []byte, processTree *proctree.ProcessTree,
	printer *output.Printer, ruleEngine *rules.Engine) {

	ev, err := decodeFileOpenEvent(data)
	if err != nil {
		log.Printf("Error decoding file open event: %v", err)
		return
	}

	filename := utils.ExtractCString(ev.Filename[:])
	if matched, rule := ruleEngine.MatchFile(filename, ev.PID, ev.CgroupID); matched && rule != nil {
		chain := processTree.GetAncestors(ev.PID)
		printer.PrintFileOpenAlert(&ev, chain, rule, filename)
	}
}

const (
	// Event sizes: type(1) + fields
	minExecEventSize     = 1 + 4 + 4 + 8 + events.TaskCommLen + events.TaskCommLen // 49 bytes
	minFileOpenEventSize = 1 + 4 + 8 + 4 + events.PathMaxLen                       // 273 bytes
)

func decodeExecEvent(data []byte) (events.ExecEvent, error) {
	if len(data) < minExecEventSize {
		return events.ExecEvent{}, fmt.Errorf("exec event payload too small: %d bytes", len(data))
	}

	var ev events.ExecEvent
	offset := 1 // Skip type byte at index 0
	ev.PID = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	ev.PPID = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	ev.CgroupID = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	copy(ev.Comm[:], data[offset:offset+16])
	offset += 16
	copy(ev.PComm[:], data[offset:offset+16])

	return ev, nil
}

func decodeFileOpenEvent(data []byte) (events.FileOpenEvent, error) {
	if len(data) < minFileOpenEventSize {
		return events.FileOpenEvent{}, fmt.Errorf("file open event too small: %d bytes", len(data))
	}

	var ev events.FileOpenEvent
	offset := 1 // Skip type byte at index 0
	ev.PID = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	ev.CgroupID = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	ev.Flags = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	copy(ev.Filename[:], data[offset:offset+events.PathMaxLen])

	return ev, nil
}

// populateMonitoredPaths extracts all monitored paths from rules and populates the BPF map
func populateMonitoredPaths(bpfMap *cebpf.Map, ruleList []rules.Rule, rulesPath string) error {
	if bpfMap == nil {
		return fmt.Errorf("monitored_paths map is nil")
	}

	// Extract unique paths from rules
	pathSet := make(map[string]struct{})

	for _, rule := range ruleList {
		// Add exact filenames
		if rule.Match.Filename != "" {
			pathSet[rule.Match.Filename] = struct{}{}
		}

		// Add file path prefixes (directory paths)
		if rule.Match.FilePath != "" {
			pathSet[rule.Match.FilePath] = struct{}{}
		}
	}

	if len(pathSet) == 0 {
		log.Printf("Warning: No file access rules found in %s", rulesPath)
		return nil
	}

	count := 0
	value := uint8(1) // Dummy value, we only care about key existence

	for path := range pathSet {
		// Convert path to fixed-size byte array (required by BPF map)
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
