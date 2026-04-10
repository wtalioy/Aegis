package proc

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
)

var cgroupPathCache sync.Map

func readCgroupPaths(pid uint32) (string, string, error) {
	file, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	var unified, hybrid string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if unified == "" {
			if after, ok := strings.CutPrefix(line, "0::"); ok {
				if trimmed := strings.TrimSpace(after); trimmed != "" {
					unified = trimmed
					if hybrid != "" {
						break
					}
					continue
				}
			}
		}

		if hybrid == "" {
			parts := strings.SplitN(line, ":", 3)
			if len(parts) == 3 && parts[1] == "" {
				if trimmed := strings.TrimSpace(parts[2]); trimmed != "" {
					hybrid = trimmed
					if unified != "" {
						break
					}
				}
			}
		}
	}

	return unified, hybrid, scanner.Err()
}

func ResolveCgroupPath(pid uint32, cgroupID uint64) string {
	if cgroupID == 0 {
		return ""
	}

	if cached, ok := cgroupPathCache.Load(cgroupID); ok {
		return cached.(string)
	}

	if unified, hybrid, err := readCgroupPaths(pid); err == nil {
		path := unified
		if path == "" {
			path = hybrid
		}
		if path != "" {
			cgroupPathCache.Store(cgroupID, path)
			return path
		}
	}

	return ""
}

func getCgroupInode(cgroupPath string) uint64 {
	if cgroupPath == "" {
		cgroupPath = "/"
	}

	cgroupMounts := []string{
		"/sys/fs/cgroup",
		"/sys/fs/cgroup/unified",
	}

	for _, mount := range cgroupMounts {
		fullPath := mount + cgroupPath
		var stat syscall.Stat_t
		if err := syscall.Stat(fullPath, &stat); err == nil {
			return stat.Ino
		}
	}

	return 0
}

func readCgroupIDAndPath(pid uint32) (uint64, string) {
	unified, hybrid, err := readCgroupPaths(pid)
	if err != nil {
		return 0, ""
	}

	path := unified
	if path == "" {
		path = hybrid
	}
	if path == "" {
		return 0, ""
	}

	return getCgroupInode(path), path
}
