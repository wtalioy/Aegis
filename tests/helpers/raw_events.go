package helpers

import (
	"encoding/binary"
	"net"
	"time"

	"aegis/internal/platform/events"
)

func RawExecSample(pid, ppid uint32, cgroupID uint64, comm, parentComm, filename, commandLine string, blocked bool) []byte {
	buf := make([]byte, events.ExecEventSize)
	encodeHeader(buf, events.EventTypeExec, pid, cgroupID, comm, blocked)
	offset := events.EventHeaderSize
	binary.LittleEndian.PutUint32(buf[offset:offset+4], ppid)
	offset += 8
	copyCString(buf[offset:offset+events.TaskCommLen], parentComm)
	offset += events.TaskCommLen
	copyCString(buf[offset:offset+events.PathMaxLen], filename)
	offset += events.PathMaxLen
	copyCString(buf[offset:offset+events.CommandLineLen], commandLine)
	return buf
}

func RawFileSample(pid uint32, cgroupID uint64, comm, filename string, flags uint32, ino, dev uint64, blocked bool) []byte {
	buf := make([]byte, events.FileOpenEventSize)
	encodeHeader(buf, events.EventTypeFileOpen, pid, cgroupID, comm, blocked)
	offset := events.EventHeaderSize
	binary.LittleEndian.PutUint64(buf[offset:offset+8], ino)
	offset += 8
	binary.LittleEndian.PutUint64(buf[offset:offset+8], dev)
	offset += 8
	binary.LittleEndian.PutUint32(buf[offset:offset+4], flags)
	offset += 8
	copyCString(buf[offset:offset+events.PathMaxLen], filename)
	return buf
}

func RawConnectSample(pid uint32, cgroupID uint64, comm, ip string, family, port uint16, blocked bool) []byte {
	buf := make([]byte, events.ConnectEventSize)
	encodeHeader(buf, events.EventTypeConnect, pid, cgroupID, comm, blocked)
	offset := events.EventHeaderSize
	if family == 2 {
		parsed := net.ParseIP(ip).To4()
		if parsed != nil {
			addr := binary.LittleEndian.Uint32(parsed)
			binary.LittleEndian.PutUint32(buf[offset:offset+4], addr)
		}
	} else if family == 10 {
		parsed := net.ParseIP(ip).To16()
		if parsed != nil {
			copy(buf[offset+8:offset+24], parsed)
		}
	}
	offset += 4
	binary.LittleEndian.PutUint16(buf[offset:offset+2], family)
	offset += 2
	binary.LittleEndian.PutUint16(buf[offset:offset+2], port)
	return buf
}

func encodeHeader(buf []byte, eventType events.EventType, pid uint32, cgroupID uint64, comm string, blocked bool) {
	offset := 0
	binary.LittleEndian.PutUint64(buf[offset:offset+8], uint64(time.Second))
	offset += 8
	binary.LittleEndian.PutUint64(buf[offset:offset+8], cgroupID)
	offset += 8
	binary.LittleEndian.PutUint32(buf[offset:offset+4], pid)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:offset+4], pid)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:offset+4], 1000)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:offset+4], 1000)
	offset += 4
	buf[offset] = byte(eventType)
	offset++
	if blocked {
		buf[offset] = 1
	}
	offset += 7
	copyCString(buf[offset:offset+events.TaskCommLen], comm)
}

func copyCString(dst []byte, value string) {
	for i := range dst {
		dst[i] = 0
	}
	copy(dst, []byte(value))
}
