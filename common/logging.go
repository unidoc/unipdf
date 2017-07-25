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
}

// Dummy Logger does nothing.
type DummyLogger struct{}

func (this DummyLogger) Error(format string, args ...interface{}) {
}

func (this DummyLogger) Warning(format string, args ...interface{}) {
}

func (this DummyLogger) Notice(format string, args ...interface{}) {
}

func (this DummyLogger) Info(format string, args ...interface{}) {
}

func (this DummyLogger) Debug(format string, args ...interface{}) {
}

func (this DummyLogger) Trace(format string, args ...interface{}) {
}

// Simple Console Logger that the tests use.
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

func (this ConsoleLogger) Error(format string, args ...interface{}) {
	if this.LogLevel >= LogLevelError {
		prefix := "[ERROR] "
		this.output(os.Stdout, prefix, format, args...)
	}
}

func (this ConsoleLogger) Warning(format string, args ...interface{}) {
	if this.LogLevel >= LogLevelWarning {
		prefix := "[WARNING] "
		this.output(os.Stdout, prefix, format, args...)
	}
}

func (this ConsoleLogger) Notice(format string, args ...interface{}) {
	if this.LogLevel >= LogLevelNotice {
		prefix := "[NOTICE] "
		this.output(os.Stdout, prefix, format, args...)
	}
}

func (this ConsoleLogger) Info(format string, args ...interface{}) {
	if this.LogLevel >= LogLevelInfo {
		prefix := "[INFO] "
		this.output(os.Stdout, prefix, format, args...)
	}
}

func (this ConsoleLogger) Debug(format string, args ...interface{}) {
	if this.LogLevel >= LogLevelDebug {
		prefix := "[DEBUG] "
		this.output(os.Stdout, prefix, format, args...)
	}
}

func (this ConsoleLogger) Trace(format string, args ...interface{}) {
	if this.LogLevel >= LogLevelTrace {
		prefix := "[TRACE] "
		this.output(os.Stdout, prefix, format, args...)
	}
}

var Log Logger = DummyLogger{}

func SetLogger(logger Logger) {
	Log = logger
}

// output writes `format`, `args` log message prefixed by the source file name, line and `prefix`
func (this ConsoleLogger) output(f *os.File, prefix string, format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}

	src := fmt.Sprintf("%s %s:%d ", prefix, file, line) + format + "\n"
	fmt.Fprintf(f, src, args...)
}
