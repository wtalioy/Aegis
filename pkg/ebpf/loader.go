package ebpf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cilium/ebpf"
)

type LSMObjects struct {
	LsmBprmCheck     *ebpf.Program `ebpf:"lsm_bprm_check"`
	LsmFileOpen      *ebpf.Program `ebpf:"lsm_file_open"`
	LsmSocketConnect *ebpf.Program `ebpf:"lsm_socket_connect"`

	Events         *ebpf.Map `ebpf:"events"`
	MonitoredFiles *ebpf.Map `ebpf:"monitored_files"`
	BlockedPorts   *ebpf.Map `ebpf:"blocked_ports"`
	PidToPpid      *ebpf.Map `ebpf:"pid_to_ppid"`
}

func LoadLSMObjects(objPath string, ringBufSize int) (*LSMObjects, error) {
	abspath, err := filepath.Abs(objPath)
	if err != nil {
		return nil, fmt.Errorf("resolve bpf path: %w", err)
	}
	if _, err := os.Stat(abspath); err != nil {
		return nil, fmt.Errorf("stat bpf object: %w", err)
	}

	spec, err := ebpf.LoadCollectionSpec(abspath)
	if err != nil {
		return nil, fmt.Errorf("load collection spec: %w", err)
	}

	if ringBufSize > 0 {
		if eventsSpec, ok := spec.Maps["events"]; ok {
			eventsSpec.MaxEntries = uint32(ringBufSize)
		}
	}

	objs := &LSMObjects{}
	if err := spec.LoadAndAssign(objs, nil); err != nil {
		return nil, fmt.Errorf("load eBPF LSM programs: %w", err)
	}

	return objs, nil
}

func (o *LSMObjects) Close() error {
	if o == nil {
		return nil
	}

	var firstErr error

	// Close programs
	firstErr = closeProgram("lsm_bprm_check", o.LsmBprmCheck, firstErr)
	firstErr = closeProgram("lsm_file_open", o.LsmFileOpen, firstErr)
	firstErr = closeProgram("lsm_socket_connect", o.LsmSocketConnect, firstErr)

	// Close maps
	firstErr = closeMap("events", o.Events, firstErr)
	firstErr = closeMap("monitored_files", o.MonitoredFiles, firstErr)
	firstErr = closeMap("blocked_ports", o.BlockedPorts, firstErr)
	firstErr = closeMap("pid_to_ppid", o.PidToPpid, firstErr)

	return firstErr
}

func closeProgram(name string, p *ebpf.Program, err error) error {
	if p == nil {
		return err
	}
	if cerr := p.Close(); cerr != nil && err == nil {
		return fmt.Errorf("close %s: %w", name, cerr)
	}
	return err
}

func closeMap(name string, m *ebpf.Map, err error) error {
	if m == nil {
		return err
	}
	if cerr := m.Close(); cerr != nil && err == nil {
		return fmt.Errorf("close %s: %w", name, cerr)
	}
	return err
}