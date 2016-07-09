/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.txt', which is part of this source code package.
 */

package common

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

var log Logger = DummyLogger{}

func SetLogger(logger Logger) {
	log = logger
}

func GetLogger() Logger {
	return log
}
