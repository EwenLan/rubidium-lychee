// Package log provides dual-output (stdout + file) leveled logging and
// structured message logging for replay. It is safe for concurrent use.
package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level controls log verbosity.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	mu   sync.Mutex
	out  io.Writer
	file *os.File
)

// Init opens a log file in dir/name and sends all subsequent output to both
// stdout and the file. The file is truncated on each run.
func Init(dir, name string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	path := filepath.Join(dir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		file.Close()
	}
	out = io.MultiWriter(os.Stdout, f)
	file = f
	return nil
}

// InitStdout enables logging to stdout only (no file). Useful for tools like
// the mock server in replay mode.
func InitStdout() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		file.Close()
		file = nil
	}
	out = os.Stdout
}

// Close closes the log file if one is open.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		file.Close()
		file = nil
	}
	out = nil
}

func write(s string) {
	mu.Lock()
	defer mu.Unlock()
	if out != nil {
		io.WriteString(out, s)
	}
}

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

func levelTag(l Level) string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "?"
	}
}

func logf(l Level, format string, args ...any) {
	write(fmt.Sprintf("%s [%s] %s\n", timestamp(), levelTag(l), fmt.Sprintf(format, args...)))
}

// Debugf logs at DEBUG level.
func Debugf(format string, args ...any) { logf(LevelDebug, format, args...) }

// Infof logs at INFO level.
func Infof(format string, args ...any) { logf(LevelInfo, format, args...) }

// Warnf logs at WARN level.
func Warnf(format string, args ...any) { logf(LevelWarn, format, args...) }

// Errorf logs at ERROR level.
func Errorf(format string, args ...any) { logf(LevelError, format, args...) }

// Fatalf logs at ERROR level, closes the log, and exits the process.
func Fatalf(format string, args ...any) {
	Errorf(format, args...)
	Close()
	os.Exit(1)
}
