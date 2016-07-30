/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package common

import (
	"fmt"
)

type Logger interface {
	Error(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Notice(format string, args ...interface{})
	Info(format string, args ...interface{})
	Debug(format string, args ...interface{})
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

// Simple Console Logger that the tests use.
type ConsoleLogger struct{}

func (this ConsoleLogger) Error(format string, args ...interface{}) {
	prefix := "[ERROR] "
	fmt.Printf(prefix+format+"\n", args...)
}

func (this ConsoleLogger) Warning(format string, args ...interface{}) {
	prefix := "[WARNING] "
	fmt.Printf(prefix+format+"\n", args...)
}

func (this ConsoleLogger) Notice(format string, args ...interface{}) {
	prefix := "[NOTICE] "
	fmt.Printf(prefix+format+"\n", args...)
}

func (this ConsoleLogger) Info(format string, args ...interface{}) {
	prefix := "[INFO] "
	fmt.Printf(prefix+format+"\n", args...)
}

func (this ConsoleLogger) Debug(format string, args ...interface{}) {
	prefix := "[DEBUG] "
	fmt.Printf(prefix+format+"\n", args...)
}

var Log Logger = DummyLogger{}

func SetLogger(logger Logger) {
	Log = logger
}
