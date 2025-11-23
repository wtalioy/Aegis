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
	"eulerguard/pkg/rules"

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

	objs, err := ebpf.LoadExecveObjects(t.opts.BPFPath)
	if err != nil {
		return err
	}
	defer objs.Close()

	tp, err := link.Tracepoint("sched", "sched_process_exec", objs.HandleExec, nil)
	if err != nil {
		return fmt.Errorf("attach tracepoint: %w", err)
	}
	defer tp.Close()

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

	printer, err := output.NewPrinter(t.opts.JSONLines, meter, t.opts.LogFile)
	if err != nil {
		return fmt.Errorf("failed to create printer: %w", err)
	}
	defer printer.Close()

	// Load rules
	loadedRules, err := rules.LoadRules(t.opts.RulesPath)
	if err != nil {
		log.Printf("Warning: failed to load rules from %s: %v", t.opts.RulesPath, err)
		log.Printf("Continuing without rules...")
		loadedRules = []rules.Rule{}
	} else {
		log.Printf("Loaded %d detection rules from %s", len(loadedRules), t.opts.RulesPath)
	}

	ruleEngine := rules.NewEngine(loadedRules)

	log.Printf("EulerGuard execve tracer ready (BPF object: %s)", t.opts.BPFPath)

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

		ev, err := decodeExecEvent(record.RawSample)
		if err != nil {
			return err
		}

		// Print the event and get the processed event
		processedEvent := printer.Print(ev)

		// Match against rules
		alerts := ruleEngine.Match(processedEvent)
		for _, alert := range alerts {
			printer.PrintAlert(alert)
		}
	}
}

func decodeExecEvent(data []byte) (events.ExecEvent, error) {
	if len(data) < 48 {
		return events.ExecEvent{}, fmt.Errorf("exec event payload too small: %d bytes", len(data))
	}

	var ev events.ExecEvent
	ev.PID = binary.LittleEndian.Uint32(data[0:4])
	ev.PPID = binary.LittleEndian.Uint32(data[4:8])
	ev.CgroupID = binary.LittleEndian.Uint64(data[8:16])
	copy(ev.Comm[:], data[16:32])
	copy(ev.PComm[:], data[32:48])

	return ev, nil
}
