package utils

import (
	"bytes"
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
		addr := event.AddrV4
		return net.IPv4(
			byte(addr),
			byte(addr>>8),
			byte(addr>>16),
			byte(addr>>24),
		).String()
	case 10: // AF_INET6 (IPv6)
		return net.IP(event.AddrV6[:]).String()
	}
	return ""
}
