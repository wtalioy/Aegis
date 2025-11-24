package utils

import (
	"bytes"
	"encoding/binary"
	"eulerguard/pkg/events"
	"net"
)

// extract a null-terminated C string from a byte array
func ExtractCString(data []byte) string {
	if idx := bytes.IndexByte(data, 0); idx != -1 {
		return string(data[:idx])
	}
	return string(data)
}

// extract the IP address from a ConnectEvent
func ExtractIP(event *events.ConnectEvent) string {
	switch event.Family {
	case 2: // AF_INET (IPv4)
		ipBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(ipBytes, event.AddrV4)
		ip := net.IPv4(ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])
		return ip.String()
	case 10: // AF_INET6 (IPv6)
		ip := net.IP(event.AddrV6[:])
		return ip.String()
	}
	return ""
}
