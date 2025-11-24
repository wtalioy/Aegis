package output

import (
	"bufio"
	"encoding/json"
	"eulerguard/pkg/metrics"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type threadSafeWriter struct {
	mu     *sync.Mutex
	writer io.Writer
}

func (w *threadSafeWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}

type Printer struct {
	jsonLines  bool
	meter      *metrics.RateMeter
	logFile    *os.File
	logWriter  *bufio.Writer
	writer     io.Writer
	stdout     io.Writer
	flushTimer *time.Ticker
	stopFlush  chan struct{}
	closeOnce  sync.Once
	mu         *sync.Mutex
}

func NewPrinter(jsonLines bool, meter *metrics.RateMeter, logPath string, bufferSize int) (*Printer, error) {
	if err := rotateLogIfNeeded(logPath); err != nil {
		log.Printf("Warning: log rotation failed: %v", err)
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logWriter := bufio.NewWriterSize(f, bufferSize)
	mu := &sync.Mutex{}
	tsLogWriter := &threadSafeWriter{
		mu:     mu,
		writer: logWriter,
	}

	p := &Printer{
		jsonLines:  jsonLines,
		meter:      meter,
		logFile:    f,
		logWriter:  logWriter,
		stdout:     os.Stdout,
		flushTimer: time.NewTicker(1 * time.Second),
		stopFlush:  make(chan struct{}),
		mu:         mu,
	}
	p.writer = io.MultiWriter(p.stdout, tsLogWriter)

	go p.flushLoop()

	log.Printf("Logging to file: %s", logPath)
	return p, nil
}

func (p *Printer) flushLoop() {
	for {
		select {
		case <-p.flushTimer.C:
			p.mu.Lock()
			if p.logWriter != nil {
				_ = p.logWriter.Flush()
			}
			p.mu.Unlock()
		case <-p.stopFlush:
			return
		}
	}
}

func (p *Printer) Close() error {
	var closeErr error

	p.closeOnce.Do(func() {
		if p.flushTimer != nil {
			p.flushTimer.Stop()
		}
		if p.stopFlush != nil {
			close(p.stopFlush)
		}

		if p.logWriter != nil {
			p.mu.Lock()
			if err := p.logWriter.Flush(); err != nil {
				log.Printf("Warning: failed to flush log buffer: %v", err)
			}
			p.mu.Unlock()
		}

		if p.logFile != nil {
			closeErr = p.logFile.Close()
		}
	})

	return closeErr
}

func (p *Printer) writeJSON(payload any, context string) {
	enc := json.NewEncoder(p.writer)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(payload); err != nil {
		log.Printf("json encode %s failed: %v", context, err)
	}
}

func (p *Printer) writeLine(format string, args ...any) {
	fmt.Fprintf(p.writer, format, args...)
}

func (p *Printer) emitColoredAlert(severity, text string) {
	fmt.Fprintf(p.stdout, "%s%s%s", getSeverityColor(severity), text, ansiReset)
	p.logText(text)
}

func (p *Printer) logText(text string) {
	p.mu.Lock()
	if p.logWriter != nil {
		fmt.Fprint(p.logWriter, text)
	}
	p.mu.Unlock()
}
