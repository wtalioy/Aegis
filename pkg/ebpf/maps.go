package ebpf

import (
	"fmt"
	"log"
	"strings"

	"aegis/pkg/events"
	"aegis/pkg/rules"

	"github.com/cilium/ebpf"
)

func PopulateMonitoredFiles(bpfMap *ebpf.Map, ruleList []rules.Rule, rulesPath string) error {
	if bpfMap == nil {
		return fmt.Errorf("monitored_files map is nil")
	}

	fileActions := make(map[string]uint8)
	for _, rule := range ruleList {
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

			action := bpfActionForRule(rule)
			fileActions[key] = mergeAction(fileActions[key], action)
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

func RepopulateMonitoredFiles(bpfMap *ebpf.Map, ruleList []rules.Rule, rulesPath string) error {
	if bpfMap == nil {
		return fmt.Errorf("monitored_files map is nil")
	}
	if err := clearMonitoredFilesMap(bpfMap); err != nil {
		return err
	}
	return PopulateMonitoredFiles(bpfMap, ruleList, rulesPath)
}

func PopulateBlockedPorts(bpfMap *ebpf.Map, ruleList []rules.Rule) error {
	if bpfMap == nil {
		return fmt.Errorf("blocked_ports map is nil")
	}

	portActions := make(map[uint16]uint8)
	for _, rule := range ruleList {
		if !rule.IsActive() {
			continue
		}

		if rule.Match.DestPort == 0 {
			continue
		}

		action := bpfActionForRule(rule)
		port := rule.Match.DestPort
		portActions[port] = mergeAction(portActions[port], action)
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

func RepopulateBlockedPorts(bpfMap *ebpf.Map, ruleList []rules.Rule) error {
	if bpfMap == nil {
		return fmt.Errorf("blocked_ports map is nil")
	}
	if err := clearBlockedPortsMap(bpfMap); err != nil {
		return err
	}
	return PopulateBlockedPorts(bpfMap, ruleList)
}

func bpfActionForRule(rule rules.Rule) uint8 {
	if rule.IsTesting() {
		return rules.BPFActionMonitor
	}
	if rule.Action == rules.ActionBlock {
		return rules.BPFActionBlock
	}
	return rules.BPFActionMonitor
}

func mergeAction(existing, proposed uint8) uint8 {
	if proposed > existing {
		return proposed
	}
	return existing
}

func clearMonitoredFilesMap(bpfMap *ebpf.Map) error {
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
	return nil
}

func clearBlockedPortsMap(bpfMap *ebpf.Map) error {
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
	return nil
}

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
