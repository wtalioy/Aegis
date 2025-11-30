package events

import (
	"encoding/binary"
	"fmt"
)

const (
	// Event sizes: type(1) + fields + blocked(1)
	MinExecEventSize     = 1 + 4 + 4 + 8 + TaskCommLen + TaskCommLen + PathMaxLen + 1 // 306 bytes
	MinFileOpenEventSize = 1 + 4 + 8 + 4 + PathMaxLen + 1                             // 274 bytes
	MinConnectEventSize  = 1 + 4 + 8 + 2 + 2 + 4 + 16 + 1                             // 38 bytes
)

func DecodeExecEvent(data []byte) (ExecEvent, error) {
	if len(data) < MinExecEventSize {
		return ExecEvent{}, fmt.Errorf("exec event payload too small: %d bytes", len(data))
	}

	var ev ExecEvent
	offset := 1 // Skip type byte at index 0
	ev.PID = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	ev.PPID = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	ev.CgroupID = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	copy(ev.Comm[:], data[offset:offset+TaskCommLen])
	offset += TaskCommLen
	copy(ev.PComm[:], data[offset:offset+TaskCommLen])
	offset += TaskCommLen
	copy(ev.Filename[:], data[offset:offset+PathMaxLen])
	ev.Blocked = data[len(data)-1]

	return ev, nil
}

func DecodeFileOpenEvent(data []byte) (FileOpenEvent, error) {
	if len(data) < MinFileOpenEventSize {
		return FileOpenEvent{}, fmt.Errorf("file open event too small: %d bytes", len(data))
	}

	var ev FileOpenEvent
	offset := 1 // Skip type byte at index 0
	ev.PID = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	ev.CgroupID = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	ev.Flags = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	copy(ev.Filename[:], data[offset:offset+PathMaxLen])
	ev.Blocked = data[len(data)-1]

	return ev, nil
}

func DecodeConnectEvent(data []byte) (ConnectEvent, error) {
	if len(data) < MinConnectEventSize {
		return ConnectEvent{}, fmt.Errorf("connect event too small: %d bytes", len(data))
	}

	var ev ConnectEvent
	offset := 1 // Skip type byte at index 0
	ev.PID = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	ev.CgroupID = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	ev.Family = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2
	ev.Port = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2
	ev.AddrV4 = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	copy(ev.AddrV6[:], data[offset:offset+16])
	ev.Blocked = data[len(data)-1]

	return ev, nil
}
