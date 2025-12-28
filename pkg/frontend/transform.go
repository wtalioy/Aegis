package frontend

import (
	"strconv"

	"aegis/pkg/apimodel"
	"aegis/pkg/events"
	"aegis/pkg/utils"
)


func ExecToFrontend(ev events.ExecEvent) apimodel.ExecEvent {
	commandLine := utils.ExtractCString(ev.CommandLine[:])
	if commandLine == "" {
		commandLine = utils.ExtractCString(ev.Filename[:])
	}
	if commandLine == "" {
		commandLine = utils.ExtractCString(ev.Hdr.Comm[:])
	}

	return apimodel.ExecEvent{
		Type:        "exec",
		Timestamp:   ev.Hdr.Timestamp().UnixMilli(),
		PID:         ev.Hdr.PID,
		PPID:        ev.PPID,
		CgroupID:    strconv.FormatUint(ev.Hdr.CgroupID, 10),
		Comm:        utils.ExtractCString(ev.Hdr.Comm[:]),
		ParentComm:  utils.ExtractCString(ev.PComm[:]),
		CommandLine: commandLine,
		Blocked:     ev.Hdr.Blocked == 1,
	}
}


func FileToFrontend(ev events.FileOpenEvent, filename string) apimodel.FileEvent {
	return apimodel.FileEvent{
		Type:      "file",
		Timestamp: ev.Hdr.Timestamp().UnixMilli(),
		PID:       ev.Hdr.PID,
		CgroupID:  strconv.FormatUint(ev.Hdr.CgroupID, 10),
		Flags:     ev.Flags,
		Ino:       ev.Ino,
		Dev:       ev.Dev,
		Filename:  filename,
		Blocked:   ev.Hdr.Blocked == 1,
	}
}


func ConnectToFrontend(ev events.ConnectEvent, addr string, processName string) apimodel.ConnectEvent {
	return apimodel.ConnectEvent{
		Type:        "connect",
		Timestamp:   ev.Hdr.Timestamp().UnixMilli(),
		PID:         ev.Hdr.PID,
		ProcessName: processName,
		CgroupID:    strconv.FormatUint(ev.Hdr.CgroupID, 10),
		Family:      ev.Family,
		Port:        ev.Port,
		Addr:        addr,
		Blocked:     ev.Hdr.Blocked == 1,
	}
}


