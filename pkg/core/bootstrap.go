package core

import (
	"fmt"
	"log"

	"aegis/pkg/config"
	"aegis/pkg/ebpf"
	"aegis/pkg/proc"
	"aegis/pkg/rules"
	"aegis/pkg/storage"
	"aegis/pkg/workload"

	cebpf "github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

type CoreComponents struct {
	EBpfObjs    *ebpf.LSMObjects
	EBpfLinks   []link.Link
	Reader      *ringbuf.Reader
	ProcessTree *proc.ProcessTree
	WorkloadReg *workload.Registry
	RuleEngine  *rules.Engine
	Rules       []rules.Rule
	Storage     *storage.Manager
	ProfileReg  *proc.ProfileRegistry
}

// Bootstrap initializes all core components in the correct order.
func Bootstrap(opts config.Options) (*CoreComponents, error) {
	// 1. Initialize process tree
	processTree := proc.NewProcessTree(
		opts.ProcessTreeMaxAge,
		opts.ProcessTreeMaxSize,
		opts.ProcessTreeMaxChainLength,
	)

	// 2. Initialize workload registry
	workloadReg := workload.NewRegistry(1000)

	// 3. Load eBPF objects
	objs, err := ebpf.LoadLSMObjects(opts.BPFPath, opts.RingBufferSize)
	if err != nil {
		return nil, fmt.Errorf("load eBPF LSM objects: %w", err)
	}

	// 4. Set PID resolver if available
	if processTree != nil && objs.PidToPpid != nil {
		processTree.SetPIDResolver(newPIDResolver(objs.PidToPpid))
	}

	// 5. Attach LSM hooks
	links, err := ebpf.AttachLSMHooks(objs)
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("attach LSM hooks: %w", err)
	}

	// 6. Create ring buffer reader
	reader, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		ebpf.CloseLinks(links)
		objs.Close()
		return nil, fmt.Errorf("create ringbuf reader: %w", err)
	}

	// 7. Load rules
	loadedRules, err := rules.LoadRules(opts.RulesPath)
	if err != nil {
		log.Printf("Warning: failed to load rules from %s: %v", opts.RulesPath, err)
		loadedRules = []rules.Rule{}
	} else {
		log.Printf("Loaded %d detection rules from %s", len(loadedRules), opts.RulesPath)
	}
	ruleEngine := rules.NewEngine(loadedRules)

	// 8. Populate BPF maps
	if err := ebpf.PopulateMonitoredFiles(objs.MonitoredFiles, loadedRules, opts.RulesPath); err != nil {
		log.Printf("Warning: failed to populate monitored files: %v", err)
	}
	if err := ebpf.PopulateBlockedPorts(objs.BlockedPorts, loadedRules); err != nil {
		log.Printf("Warning: failed to populate blocked ports: %v", err)
	}

	// 9. Initialize storage manager
	storageCapacity := config.DefaultRecentEventsCapacity
	storageManager := storage.NewManager(storageCapacity, 1000)

	// 10. Initialize profile registry
	profileReg := proc.NewProfileRegistry()

	return &CoreComponents{
		EBpfObjs:    objs,
		EBpfLinks:   links,
		Reader:      reader,
		ProcessTree: processTree,
		WorkloadReg: workloadReg,
		RuleEngine:  ruleEngine,
		Rules:       loadedRules,
		Storage:     storageManager,
		ProfileReg:  profileReg,
	}, nil
}

// ReloadRules reloads rules and updates BPF maps.
func (c *CoreComponents) ReloadRules(rulesPath string) error {
	newRules, err := rules.LoadRules(rulesPath)
	if err != nil {
		return fmt.Errorf("load rules: %w", err)
	}

	c.Rules = newRules
	c.RuleEngine = rules.NewEngine(newRules)

	if c.EBpfObjs != nil {
		if c.EBpfObjs.MonitoredFiles != nil {
			if err := ebpf.RepopulateMonitoredFiles(c.EBpfObjs.MonitoredFiles, newRules, rulesPath); err != nil {
				return fmt.Errorf("failed to repopulate monitored files: %w", err)
			}
		}
		if c.EBpfObjs.BlockedPorts != nil {
			if err := ebpf.RepopulateBlockedPorts(c.EBpfObjs.BlockedPorts, newRules); err != nil {
				return fmt.Errorf("failed to repopulate blocked ports: %w", err)
			}
		}
	}

	log.Printf("Rules reloaded: %d rules from %s", len(newRules), rulesPath)
	return nil
}

func (c *CoreComponents) Close() error {
	var firstErr error

	if c.Reader != nil {
		if err := c.Reader.Close(); err != nil {
			firstErr = err
		}
	}

	ebpf.CloseLinks(c.EBpfLinks)

	if c.EBpfObjs != nil {
		if err := c.EBpfObjs.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func newPIDResolver(m *cebpf.Map) proc.PIDResolver {
	if m == nil {
		return nil
	}

	return func(pid uint32) (uint32, bool) {
		var parent uint32
		if err := m.Lookup(&pid, &parent); err != nil {
			return 0, false
		}
		return parent, true
	}
}
