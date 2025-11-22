package output

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	maxLogSize = 100 * 1024 * 1024 // 100 MB
	maxBackups = 5
)

func rotateLogIfNeeded(logPath string) error {
	if logPath == "" {
		return nil
	}

	info, err := os.Stat(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if info.Size() < maxLogSize {
		return nil
	}

	return rotateLog(logPath)
}

func rotateLog(logPath string) error {
	timestamp := time.Now().Format("20251122-150405")
	dir := filepath.Dir(logPath)
	base := filepath.Base(logPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	backupPath := filepath.Join(dir, fmt.Sprintf("%s-%s%s", name, timestamp, ext))

	if err := os.Rename(logPath, backupPath); err != nil {
		return fmt.Errorf("failed to rotate log: %w", err)
	}

	go cleanupOldLogs(dir, name, ext)

	return nil
}

func cleanupOldLogs(dir, baseName, ext string) {
	pattern := filepath.Join(dir, fmt.Sprintf("%s-*%s", baseName, ext))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	if len(matches) <= maxBackups {
		return
	}

	for i := 0; i < len(matches)-maxBackups; i++ {
		_ = os.Remove(matches[i])
	}
}
