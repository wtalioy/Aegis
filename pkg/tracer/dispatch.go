package tracer

import (
	"log"

	"aegis/pkg/events"
	"aegis/pkg/proc"
	"aegis/pkg/storage"
	"aegis/pkg/utils"
	"aegis/pkg/workload"
)

// DispatchEvent decodes and dispatches an event to handlers.
func DispatchEvent(data []byte, handlers *events.HandlerChain, processTree *proc.ProcessTree, registry *workload.Registry, storageMgr *storage.Manager, profileReg *proc.ProfileRegistry) {
	if len(data) < events.EventHeaderSize {
		return
	}
	// Event type is at offset 32 in the header (after timestamp, cgroup_id, pid, tid, uid, gid)
	// 8+8+4+4+4+4 = 32 bytes
	eventType := events.EventType(data[32])

	switch eventType {
	case events.EventTypeExec:
		ev, err := events.DecodeExecEvent(data)
		if err != nil {
			log.Printf("Error decoding exec event: %v", err)
			return
		}
		processTree.AddProcess(ev.Hdr.PID, ev.PPID, ev.Hdr.CgroupID, utils.ExtractCString(ev.Hdr.Comm[:]))
		if registry != nil {
			cgroupPath := proc.ResolveCgroupPath(ev.Hdr.PID, ev.Hdr.CgroupID)
			registry.RecordExec(ev.Hdr.CgroupID, cgroupPath)
		}
		// Store event
		if storageMgr != nil {
			// Store the value event; snapshot/AI code handles both value and pointer forms.
			storeEvent := storage.EventFromBackend(events.EventTypeExec, ev.Hdr.Timestamp(), ev)
			_ = storageMgr.Append(storeEvent)
		}
		// Update profile
		if profileReg != nil {
			profileReg.RecordExec(ev.Hdr.PID)
		}
		handlers.HandleExec(ev)

	case events.EventTypeFileOpen:
		ev, err := events.DecodeFileOpenEvent(data)
		if err != nil {
			log.Printf("Error decoding file open event: %v", err)
			return
		}
		if registry != nil {
			cgroupPath := proc.ResolveCgroupPath(ev.Hdr.PID, ev.Hdr.CgroupID)
			registry.RecordFile(ev.Hdr.CgroupID, cgroupPath)
		}
		// Store event
		if storageMgr != nil {
			// Store the value event; snapshot/AI code handles both value and pointer forms.
			storeEvent := storage.EventFromBackend(events.EventTypeFileOpen, ev.Hdr.Timestamp(), ev)
			_ = storageMgr.Append(storeEvent)
		}
		// Update profile
		if profileReg != nil {
			profileReg.RecordFileOpen(ev.Hdr.PID)
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
			cgroupPath := proc.ResolveCgroupPath(ev.Hdr.PID, ev.Hdr.CgroupID)
			registry.RecordConnect(ev.Hdr.CgroupID, cgroupPath)
		}
		// Store event
		if storageMgr != nil {
			// Store the value event; snapshot/AI code handles both value and pointer forms.
			storeEvent := storage.EventFromBackend(events.EventTypeConnect, ev.Hdr.Timestamp(), ev)
			_ = storageMgr.Append(storeEvent)
		}
		// Update profile
		if profileReg != nil {
			profileReg.RecordConnect(ev.Hdr.PID)
		}
		handlers.HandleConnect(ev)
	}
}
