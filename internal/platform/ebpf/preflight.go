package ebpf

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

func ensureBPFLSMEnabled() error {
	enabled, source, lsms, known := detectBPFLSMState()
	if !known || enabled {
		return nil
	}

	current := strings.Join(lsms, ",")
	suggested := appendBPFToLSMList(lsms)
	if suggested == "" {
		suggested = "bpf"
	}

	return fmt.Errorf(
		"BPF LSM is not enabled on this kernel (%s=%q). Aegis uses only lsm/* hooks for exec/file/network telemetry, so no runtime events can be captured until `bpf` is added to the active LSM list (for example `lsm=%s`)",
		source,
		current,
		suggested,
	)
}

func detectBPFLSMState() (enabled bool, source string, lsms []string, known bool) {
	if active, ok := readActiveLSMs("/sys/kernel/security/lsm"); ok {
		return containsLSM(active, "bpf"), "/sys/kernel/security/lsm", active, true
	}

	if cmdline, err := os.ReadFile("/proc/cmdline"); err == nil {
		if configured, ok := extractLSMListFromCmdline(string(cmdline)); ok {
			return containsLSM(configured, "bpf"), "/proc/cmdline", configured, true
		}
	}

	if configured, source, ok := readConfiguredLSMs(); ok {
		return containsLSM(configured, "bpf"), source, configured, true
	}

	return false, "", nil, false
}

func readActiveLSMs(path string) ([]string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	list := splitLSMList(string(data))
	return list, len(list) > 0
}

func readConfiguredLSMs() ([]string, string, bool) {
	if configured, ok := readConfiguredLSMsFromProc(); ok {
		return configured, "/proc/config.gz", true
	}

	release, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return nil, "", false
	}

	path := "/boot/config-" + strings.TrimSpace(string(release))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", false
	}

	configured, ok := extractConfiguredLSMs(string(data))
	if !ok {
		return nil, "", false
	}

	return configured, path, true
}

func readConfiguredLSMsFromProc() ([]string, bool) {
	file, err := os.Open("/proc/config.gz")
	if err != nil {
		return nil, false
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, false
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, false
	}

	return extractConfiguredLSMs(string(data))
}

func extractLSMListFromCmdline(cmdline string) ([]string, bool) {
	scanner := bufio.NewScanner(strings.NewReader(cmdline))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		token := scanner.Text()
		if !strings.HasPrefix(token, "lsm=") {
			continue
		}
		value := strings.TrimPrefix(token, "lsm=")
		list := splitLSMList(value)
		return list, len(list) > 0
	}

	return nil, false
}

func extractConfiguredLSMs(config string) ([]string, bool) {
	scanner := bufio.NewScanner(strings.NewReader(config))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "CONFIG_LSM=") {
			continue
		}
		value := strings.TrimPrefix(line, "CONFIG_LSM=")
		list := splitLSMList(value)
		return list, len(list) > 0
	}

	return nil, false
}

func splitLSMList(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"'")
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result = append(result, part)
	}
	return result
}

func containsLSM(lsms []string, target string) bool {
	for _, lsm := range lsms {
		if lsm == target {
			return true
		}
	}
	return false
}

func appendBPFToLSMList(lsms []string) string {
	if containsLSM(lsms, "bpf") {
		return strings.Join(lsms, ",")
	}
	if len(lsms) == 0 {
		return "bpf"
	}
	extended := append(append([]string(nil), lsms...), "bpf")
	return strings.Join(extended, ",")
}
