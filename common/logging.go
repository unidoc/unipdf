/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Logger is the interface used for logging in the unipdf package.
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

// Defines log level enum where the most important logs have the lowest values.
// I.e. level error = 0 and level trace = 5
const (
	LogLevelTrace   LogLevel = 5
	LogLevelDebug   LogLevel = 4
	LogLevelInfo    LogLevel = 3
	LogLevelNotice  LogLevel = 2
	LogLevelWarning LogLevel = 1
	LogLevelError   LogLevel = 0
)

// ConsoleLogger is a logger that writes logs to the 'os.Stdout'
type ConsoleLogger struct {
	LogLevel LogLevel
}

// NewConsoleLogger creates new console logger.
func NewConsoleLogger(logLevel LogLevel) *ConsoleLogger {
	return &ConsoleLogger{LogLevel: logLevel}
}

// IsLogLevel returns true if log level is greater or equal than `level`.
// Can be used to avoid resource intensive calls to loggers.
func (l ConsoleLogger) IsLogLevel(level LogLevel) bool {
	return l.LogLevel >= level
}

// Error logs error message.
func (l ConsoleLogger) Error(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelError {
		prefix := "[ERROR] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

// Warning logs warning message.
func (l ConsoleLogger) Warning(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelWarning {
		prefix := "[WARNING] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

// Notice logs notice message.
func (l ConsoleLogger) Notice(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelNotice {
		prefix := "[NOTICE] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

// Info logs info message.
func (l ConsoleLogger) Info(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelInfo {
		prefix := "[INFO] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

// Debug logs debug message.
func (l ConsoleLogger) Debug(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelDebug {
		prefix := "[DEBUG] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

// Trace logs trace message.
func (l ConsoleLogger) Trace(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelTrace {
		prefix := "[TRACE] "
		l.output(os.Stdout, prefix, format, args...)
	}
}

// output writes `format`, `args` log message prefixed by the source file name, line and `prefix`
func (l ConsoleLogger) output(f io.Writer, prefix string, format string, args ...interface{}) {
	logToWriter(f, prefix, format, args...)
}

var Log Logger = DummyLogger{}

// SetLogger sets 'logger' to be used by the unidoc unipdf library.
func SetLogger(logger Logger) {
	Log = logger
}

// WriterLogger is the logger that writes data to the Output writer
type WriterLogger struct {
	LogLevel LogLevel
	Output   io.Writer
}

// NewWriterLogger creates new 'writer' logger.
func NewWriterLogger(logLevel LogLevel, writer io.Writer) *WriterLogger {
	logger := WriterLogger{
		Output:   writer,
		LogLevel: logLevel,
	}
	return &logger
}

// IsLogLevel returns true if log level is greater or equal than `level`.
// Can be used to avoid resource intensive calls to loggers.
func (l WriterLogger) IsLogLevel(level LogLevel) bool {
	return l.LogLevel >= level
}

// Error logs error message.
func (l WriterLogger) Error(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelError {
		prefix := "[ERROR] "
		l.logToWriter(l.Output, prefix, format, args...)
	}
}

// Warning logs warning message.
func (l WriterLogger) Warning(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelWarning {
		prefix := "[WARNING] "
		l.logToWriter(l.Output, prefix, format, args...)
	}
}

// Notice logs notice message.
func (l WriterLogger) Notice(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelNotice {
		prefix := "[NOTICE] "
		l.logToWriter(l.Output, prefix, format, args...)
	}
}

// Info logs info message.
func (l WriterLogger) Info(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelInfo {
		prefix := "[INFO] "
		l.logToWriter(l.Output, prefix, format, args...)
	}
}

// Debug logs debug message.
func (l WriterLogger) Debug(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelDebug {
		prefix := "[DEBUG] "
		l.logToWriter(l.Output, prefix, format, args...)
	}
}

// Trace logs trace message.
func (l WriterLogger) Trace(format string, args ...interface{}) {
	if l.LogLevel >= LogLevelTrace {
		prefix := "[TRACE] "
		l.logToWriter(l.Output, prefix, format, args...)
	}
}

// logToWriter writes `format`, `args` log message prefixed by the source file name, line and `prefix`
func (l WriterLogger) logToWriter(f io.Writer, prefix string, format string, args ...interface{}) {
	logToWriter(f, prefix, format, args)
}

func logToWriter(f io.Writer, prefix string, format string, args ...interface{}) {
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
