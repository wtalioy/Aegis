package ebpf

import (
	"fmt"

	internalconfig "aegis/internal/platform/config"
	"aegis/internal/policy"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

type Resources struct {
	Objects *LSMObjects
	Links   []link.Link
	Reader  *ringbuf.Reader
}

func Load(cfg internalconfig.Config) (*Resources, error) {
	if err := ensureBPFLSMEnabled(); err != nil {
		return nil, err
	}

	objects, err := LoadLSMObjects(cfg.Kernel.BPFPath, cfg.Kernel.RingBufferSize)
	if err != nil {
		return nil, fmt.Errorf("load eBPF objects: %w", err)
	}

	links, err := AttachLSMHooks(objects)
	if err != nil {
		objects.Close()
		return nil, fmt.Errorf("attach eBPF hooks: %w", err)
	}

	reader, err := ringbuf.NewReader(objects.Events)
	if err != nil {
		CloseLinks(links)
		objects.Close()
		return nil, fmt.Errorf("create ring buffer reader: %w", err)
	}

	return &Resources{
		Objects: objects,
		Links:   links,
		Reader:  reader,
	}, nil
}

func (r *Resources) Close() error {
	if r == nil {
		return nil
	}

	var firstErr error
	if r.Reader != nil {
		if err := r.Reader.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	CloseLinks(r.Links)
	if r.Objects != nil {
		if err := r.Objects.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type KernelSync struct {
	resources *Resources
	rulesPath string
}

func NewKernelSync(resources *Resources, rulesPath string) *KernelSync {
	return &KernelSync{
		resources: resources,
		rulesPath: rulesPath,
	}
}

func (k *KernelSync) SyncRules(ruleList []policy.Rule) error {
	if k == nil || k.resources == nil || k.resources.Objects == nil {
		return nil
	}

	if monitored := k.resources.Objects.MonitoredFiles; monitored != nil {
		if err := RepopulateMonitoredFiles(monitored, ruleList, k.rulesPath); err != nil {
			return err
		}
	}
	if blocked := k.resources.Objects.BlockedPorts; blocked != nil {
		if err := RepopulateBlockedPorts(blocked, ruleList); err != nil {
			return err
		}
	}
	return nil
}
