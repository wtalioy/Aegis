package ebpf

import (
	"fmt"
	"log"
	"strings"

	"aegis/pkg/events"
	"aegis/pkg/rules"

	"github.com/cilium/ebpf"
)

// PopulateMonitoredFiles populates the monitored_files BPF map from rules.
func PopulateMonitoredFiles(bpfMap *ebpf.Map, ruleList []rules.Rule, rulesPath string) error {
	if bpfMap == nil {
		return fmt.Errorf("monitored_files map is nil")
	}

	fileActions := make(map[string]uint8)
	for _, rule := range ruleList {
		// Only include active rules (testing or production), exclude draft rules
		if !rule.IsActive() {
			continue
		}

		paths := rule.Match.ExactPathKeys()
		if len(paths) == 0 {
			continue
		}

		for _, path := range paths {
			key := extractParentFilename(path)
			if key == "" {
				continue
			}

			var action uint8
			// Testing rules should always use monitor action (don't block)
			// Production rules use block or monitor based on their action
			if rule.IsTesting() {
				action = rules.BPFActionMonitor // Always monitor for testing rules
			} else if rule.Action == rules.ActionBlock {
				action = rules.BPFActionBlock
			} else {
				action = rules.BPFActionMonitor
			}

			// For testing rules, use monitor even if a production rule wants to block
			// (testing takes precedence - we want to observe, not block)
			if existing, ok := fileActions[key]; !ok || (rule.IsTesting() && existing == rules.BPFActionBlock) {
				fileActions[key] = action
			} else if !rule.IsTesting() && action > existing {
				fileActions[key] = action
			}
		}
	}

	if len(fileActions) == 0 {
		log.Printf("Warning: No file access rules found in %s", rulesPath)
		return nil
	}

	countMonitor := 0
	countBlock := 0
	for filename, action := range fileActions {
		key := make([]byte, events.PathMaxLen)
		copy(key, []byte(filename))
		if err := bpfMap.Put(key, action); err != nil {
			return fmt.Errorf("add file %q to BPF map: %w", filename, err)
		}
		if action == rules.BPFActionBlock {
			countBlock++
		} else {
			countMonitor++
		}
	}

	log.Printf("Populated BPF map with %d monitored files (%d block, %d monitor)",
		len(fileActions), countBlock, countMonitor)

	return nil
}

// RepopulateMonitoredFiles clears and repopulates the monitored_files BPF map.
func RepopulateMonitoredFiles(bpfMap *ebpf.Map, ruleList []rules.Rule, rulesPath string) error {
	if bpfMap == nil {
		return fmt.Errorf("monitored_files map is nil")
	}

	var key [events.PathMaxLen]byte
	var val uint8
	iter := bpfMap.Iterate()
	keysToDelete := make([][]byte, 0)
	for iter.Next(&key, &val) {
		keyCopy := make([]byte, events.PathMaxLen)
		copy(keyCopy, key[:])
		keysToDelete = append(keysToDelete, keyCopy)
	}
	for _, k := range keysToDelete {
		_ = bpfMap.Delete(k)
	}

	return PopulateMonitoredFiles(bpfMap, ruleList, rulesPath)
}

// PopulateBlockedPorts populates the blocked_ports BPF map from rules.
func PopulateBlockedPorts(bpfMap *ebpf.Map, ruleList []rules.Rule) error {
	if bpfMap == nil {
		return fmt.Errorf("blocked_ports map is nil")
	}

	portActions := make(map[uint16]uint8)
	for _, rule := range ruleList {
		// Only include active rules (testing or production), exclude draft rules
		if !rule.IsActive() {
			continue
		}

		if rule.Match.DestPort == 0 {
			continue
		}

		var action uint8
		// Testing rules should always use monitor action (don't block)
		// Production rules use block or monitor based on their action
		if rule.IsTesting() {
			action = rules.BPFActionMonitor // Always monitor for testing rules
		} else if rule.Action == rules.ActionBlock {
			action = rules.BPFActionBlock
		} else {
			action = rules.BPFActionMonitor
		}

		port := rule.Match.DestPort
		// For testing rules, use monitor even if a production rule wants to block
		// (testing takes precedence - we want to observe, not block)
		if existing, ok := portActions[port]; !ok || (rule.IsTesting() && existing == rules.BPFActionBlock) {
			portActions[port] = action
		} else if !rule.IsTesting() && action > existing {
			portActions[port] = action
		}
	}

	if len(portActions) == 0 {
		return nil
	}

	countMonitor := 0
	countBlock := 0
	for port, action := range portActions {
		if err := bpfMap.Put(port, action); err != nil {
			return fmt.Errorf("add port %d to BPF map: %w", port, err)
		}
		if action == rules.BPFActionBlock {
			countBlock++
		} else {
			countMonitor++
		}
	}

	log.Printf("Populated BPF map with %d monitored ports (%d block, %d monitor)",
		len(portActions), countBlock, countMonitor)
	return nil
}

// RepopulateBlockedPorts clears and repopulates the blocked_ports BPF map.
func RepopulateBlockedPorts(bpfMap *ebpf.Map, ruleList []rules.Rule) error {
	if bpfMap == nil {
		return fmt.Errorf("blocked_ports map is nil")
	}

	var key uint16
	var val uint8
	iter := bpfMap.Iterate()
	keysToDelete := make([]uint16, 0)
	for iter.Next(&key, &val) {
		keysToDelete = append(keysToDelete, key)
	}
	for _, k := range keysToDelete {
		_ = bpfMap.Delete(k)
	}

	return PopulateBlockedPorts(bpfMap, ruleList)
}

// extractParentFilename extracts the parent/filename format from a full path.
func extractParentFilename(path string) string {
	for len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	if path == "" {
		return ""
	}

	segments := strings.FieldsFunc(path, func(r rune) bool { return r == '/' })
	if len(segments) == 0 {
		return ""
	}

	start := max(len(segments)-3, 0)
	return strings.Join(segments[start:], "/")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
