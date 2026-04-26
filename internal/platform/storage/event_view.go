package storage

import (
	"aegis/internal/platform/events"
	"aegis/internal/shared/utils"
)

type EventView struct {
	Type        events.EventType
	PID         uint32
	PPID        uint32
	CgroupID    uint64
	ProcessName string
	ParentName  string
	CommandLine string
	Filename    string
	Flags       uint32
	Ino         uint64
	Dev         uint64
	Family      uint16
	Port        uint16
	Address     string
	Blocked     bool
}

func ExecPayload(event *Event) (events.ExecEvent, bool) {
	if event == nil {
		return events.ExecEvent{}, false
	}
	switch data := event.Data.(type) {
	case events.ExecEvent:
		return data, true
	case *events.ExecEvent:
		if data == nil {
			return events.ExecEvent{}, false
		}
		return *data, true
	default:
		return events.ExecEvent{}, false
	}
}

func FileOpenPayload(event *Event) (events.FileOpenEvent, bool) {
	if event == nil {
		return events.FileOpenEvent{}, false
	}
	switch data := event.Data.(type) {
	case events.FileOpenEvent:
		return data, true
	case *events.FileOpenEvent:
		if data == nil {
			return events.FileOpenEvent{}, false
		}
		return *data, true
	default:
		return events.FileOpenEvent{}, false
	}
}

func ConnectPayload(event *Event) (events.ConnectEvent, bool) {
	if event == nil {
		return events.ConnectEvent{}, false
	}
	switch data := event.Data.(type) {
	case events.ConnectEvent:
		return data, true
	case *events.ConnectEvent:
		if data == nil {
			return events.ConnectEvent{}, false
		}
		return *data, true
	default:
		return events.ConnectEvent{}, false
	}
}

func View(event *Event) (EventView, bool) {
	if execEvent, ok := ExecPayload(event); ok {
		return EventView{
			Type:        events.EventTypeExec,
			PID:         execEvent.Hdr.PID,
			PPID:        execEvent.PPID,
			CgroupID:    execEvent.Hdr.CgroupID,
			ProcessName: utils.ExtractCString(execEvent.Hdr.Comm[:]),
			ParentName:  utils.ExtractCString(execEvent.PComm[:]),
			CommandLine: utils.ExtractCString(execEvent.CommandLine[:]),
			Filename:    utils.ExtractCString(execEvent.Filename[:]),
			Blocked:     execEvent.Hdr.Blocked == 1,
		}, true
	}
	if fileEvent, ok := FileOpenPayload(event); ok {
		return EventView{
			Type:        events.EventTypeFileOpen,
			PID:         fileEvent.Hdr.PID,
			CgroupID:    fileEvent.Hdr.CgroupID,
			ProcessName: utils.ExtractCString(fileEvent.Hdr.Comm[:]),
			Filename:    utils.ExtractCString(fileEvent.Filename[:]),
			Flags:       fileEvent.Flags,
			Ino:         fileEvent.Ino,
			Dev:         fileEvent.Dev,
			Blocked:     fileEvent.Hdr.Blocked == 1,
		}, true
	}
	if connectEvent, ok := ConnectPayload(event); ok {
		return EventView{
			Type:        events.EventTypeConnect,
			PID:         connectEvent.Hdr.PID,
			CgroupID:    connectEvent.Hdr.CgroupID,
			ProcessName: utils.ExtractCString(connectEvent.Hdr.Comm[:]),
			Family:      connectEvent.Family,
			Port:        connectEvent.Port,
			Address:     utils.ExtractIP(&connectEvent),
			Blocked:     connectEvent.Hdr.Blocked == 1,
		}, true
	}
	return EventView{}, false
}
