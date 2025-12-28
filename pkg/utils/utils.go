package utils

import (
	"slices"
	"bytes"
	"net"
	"path/filepath"
	"strings"
	"regexp"

	"aegis/pkg/events"
)

var (
	numericSegmentRe  = regexp.MustCompile(`^\d+$`)
	hexSegmentRe      = regexp.MustCompile(`^(?:0x)?[0-9a-fA-F]{6,}$`)
	ephemeralPrefixes = []string{
		"/tmp/",
		"/var/tmp/",
		"/dev/shm/",
	}
	sensitivePrefixes = []string{
		"/etc/",
		"/bin/",
		"/sbin/",
		"/usr/bin/",
		"/usr/sbin/",
	}
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

// normalize a filename by cleaning the path and stripping leading slashes
func NormalizeFilename(path string) string {
	if path == "" {
		return ""
	}
	clean := filepath.Clean(path)
	if clean == "." {
		return ""
	}
	return strings.TrimLeft(clean, "/")
}

// return canonical and relative forms of a path for matching.
func PathVariants(path string) []string {
	if path == "" {
		return nil
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == "" {
		return nil
	}

	var variants []string
	add := func(val string) {
		if val == "" {
			return
		}
		if slices.Contains(variants, val) {
			return
		}
		variants = append(variants, val)
	}

	add(clean)
	trimmed := strings.TrimLeft(clean, "/")
	add(trimmed)
	return variants
}

func SimplifyPath(raw string) string {
	path := strings.TrimSpace(raw)
	if path == "" {
		return ""
	}

	cleaned := filepath.Clean(path)
	if cleaned == "." {
		return path
	}

	for _, prefix := range sensitivePrefixes {
		base := strings.TrimSuffix(prefix, "/")
		if cleaned == base || strings.HasPrefix(cleaned, prefix) {
			return cleaned
		}
	}

	for _, prefix := range ephemeralPrefixes {
		base := strings.TrimSuffix(prefix, "/")
		if strings.HasPrefix(cleaned, prefix) || cleaned == base {
			return base + "/*"
		}
	}

	segments := strings.Split(cleaned, "/")
	for i, segment := range segments {
		if segment == "" {
			continue
		}
		if numericSegmentRe.MatchString(segment) || hexSegmentRe.MatchString(segment) {
			segments[i] = "*"
		}
	}

	simplified := strings.Join(segments, "/")
	if simplified == "" {
		return "/"
	}
	return simplified
}
