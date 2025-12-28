package tracer

import (
	"errors"
	"fmt"
	"syscall"

	"aegis/pkg/events"
	"aegis/pkg/proc"
	"aegis/pkg/storage"
	"aegis/pkg/workload"

	"github.com/cilium/ebpf/ringbuf"
)

// EventLoop reads events from the ring buffer and dispatches them.
func EventLoop(reader *ringbuf.Reader, handlers *events.HandlerChain, processTree *proc.ProcessTree, registry *workload.Registry, storageMgr *storage.Manager, profileReg *proc.ProfileRegistry) error {
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

		DispatchEvent(record.RawSample, handlers, processTree, registry, storageMgr, profileReg)
	}
}
