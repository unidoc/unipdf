/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Logger interface {
	Error(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Notice(format string, args ...interface{})
	Info(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Trace(format string, args ...interface{})
	IsLogLevel(level LogLevel) bool
}

// DummyLogger does nothing.
type DummyLogger struct{}

// Error does nothing for dummy logger.
func (DummyLogger) Error(format string, args ...interface{}) {
}

// Warning does nothing for dummy logger.
func (DummyLogger) Warning(format string, args ...interface{}) {
}

// Notice does nothing for dummy logger.
func (DummyLogger) Notice(format string, args ...interface{}) {
}

// Info does nothing for dummy logger.
func (DummyLogger) Info(format string, args ...interface{}) {
}

// Debug does nothing for dummy logger.
func (DummyLogger) Debug(format string, args ...interface{}) {
}

// Trace does nothing for dummy logger.
func (DummyLogger) Trace(format string, args ...interface{}) {
}

// IsLogLevel returns true from dummy logger.
func (DummyLogger) IsLogLevel(level LogLevel) bool {
	return true
}

// LogLevel is the verbosity level for logging.
type LogLevel int

const (
	LogLevelTrace   LogLevel = 5
	LogLevelDebug   LogLevel = 4
	LogLevelInfo    LogLevel = 3
	LogLevelNotice  LogLevel = 2
	LogLevelWarning LogLevel = 1
	LogLevelError   LogLevel = 0
)

type ConsoleLogger struct {
	LogLevel LogLevel
}

func NewConsoleLogger(logLevel LogLevel) *ConsoleLogger {
	logger := ConsoleLogger{}
	logger.LogLevel = logLevel
	return &logger
}

// IsLogLevel returns true if log level is greater or equal than `level`.
// Can be used to avoid resource intensive calls to loggers.
func (l ConsoleLogger) IsLogLevel(level LogLevel) bool {
	return l.LogLevel >= level
}

func (l ConsoleLogger) Error(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelError {
		prefix := "[ERROR] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

func (l ConsoleLogger) Warning(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelWarning {
		prefix := "[WARNING] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

func (l ConsoleLogger) Notice(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelNotice {
		prefix := "[NOTICE] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

func (l ConsoleLogger) Info(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelInfo {
		prefix := "[INFO] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

func (l ConsoleLogger) Debug(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelDebug {
		prefix := "[DEBUG] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

func (l ConsoleLogger) Trace(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelTrace {
		prefix := "[TRACE] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

var Log Logger = DummyLogger{}

func SetLogger(logger Logger) {
	Log = logger
}

// output writes `format`, `args` log message prefixed by the source file name, line and `prefix`
func (l ConsoleLogger) output(f *os.File, prefix string, format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}

	src := fmt.Sprintf("%s %s:%d ", prefix, file, line) + format + "\n"
	fmt.Fprintf(f, src, args...)
}
