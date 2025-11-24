package ebpf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cilium/ebpf"
)

type ExecveObjects struct {
	HandleExec        *ebpf.Program `ebpf:"handle_exec"`
	TracepointOpenat  *ebpf.Program `ebpf:"tracepoint_openat"`
	TracepointConnect *ebpf.Program `ebpf:"tracepoint_connect"`
	Events            *ebpf.Map     `ebpf:"events"`
	MonitoredPaths    *ebpf.Map     `ebpf:"monitored_paths"`
}

func LoadExecveObjects(objPath string, ringBufSize int) (*ExecveObjects, error) {
	absPath, err := filepath.Abs(objPath)
	if err != nil {
		return nil, fmt.Errorf("resolve bpf path: %w", err)
	}
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("stat bpf object: %w", err)
	}

	spec, err := ebpf.LoadCollectionSpec(absPath)
	if err != nil {
		return nil, fmt.Errorf("load collection spec: %w", err)
	}

	if ringBufSize > 0 {
		if eventsSpec, ok := spec.Maps["events"]; ok {
			eventsSpec.MaxEntries = uint32(ringBufSize)
		}
	}

	objs := &ExecveObjects{}
	if err := spec.LoadAndAssign(objs, nil); err != nil {
		return nil, fmt.Errorf("load eBPF programs: %w", err)
	}

	return objs, nil
}

func (o *ExecveObjects) Close() error {
	if o == nil {
		return nil
	}

	var firstErr error
	if o.HandleExec != nil {
		if err := o.HandleExec.Close(); err != nil {
			firstErr = fmt.Errorf("close handle_exec: %w", err)
		}
	}
	if o.TracepointOpenat != nil {
		if err := o.TracepointOpenat.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close tracepoint_openat: %w", err)
		}
	}
	if o.TracepointConnect != nil {
		if err := o.TracepointConnect.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close tracepoint_connect: %w", err)
		}
	}
	if o.Events != nil {
		if err := o.Events.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close events map: %w", err)
		}
	}
	if o.MonitoredPaths != nil {
		if err := o.MonitoredPaths.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close monitored_paths map: %w", err)
		}
	}
	return firstErr
}
