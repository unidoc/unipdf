/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package common

import (
	"fmt"
	"path/filepath"
	"runtime"
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
	this.output("[ERROR] ", format, args...)
}

func (this ConsoleLogger) Warning(format string, args ...interface{}) {
	this.output("[WARNING] ", format, args...)
}

func (this ConsoleLogger) Notice(format string, args ...interface{}) {
	this.output("[NOTICE] ", format, args...)
}

func (this ConsoleLogger) Info(format string, args ...interface{}) {
	this.output("[INFO] ", format, args...)
}

func (this ConsoleLogger) Debug(format string, args ...interface{}) {
	this.output("[DEBUG] ", format, args...)
}

func (this ConsoleLogger) output(prefix, format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}
	src := fmt.Sprintf("%s:%d %s", file, line, prefix) + format + "\n"
	fmt.Printf(src, args...)
}

var Log Logger = DummyLogger{}

func SetLogger(logger Logger) {
	Log = logger
	fmt.Printf("SetLogger: logger=%+v\n", logger)
}
