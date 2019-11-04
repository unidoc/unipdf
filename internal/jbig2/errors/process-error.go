/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package errors

import (
	"fmt"
)

type processError struct {
	header  string
	process string
	message string
	wrapped error
}

func (p *processError) Error() string {
	var message string
	if p.header != "" {
		message = p.header
	}
	message += "Process: " + p.process
	if p.message != "" {
		message += " Message: " + p.message
	}

	if p.wrapped != nil {
		message += ". " + p.wrapped.Error()
	}
	return message

}

// Error returns an error wrapped with provided 'process' and with given 'message'.
func Error(processName, message string) error {
	return newProcessError(message, processName)
}

// Errorf returns an error with provided message, arguments and process name.
func Errorf(processName, message string, arguments ...interface{}) error {
	return newProcessError(fmt.Sprintf(message, arguments...), processName)
}

func newProcessError(message, processName string) *processError {
	return &processError{header: "[JBIG2]", message: message, process: processName}
}

// Wrap wraps the error with the message and provided process.
func Wrap(err error, processName, message string) error {
	if perror, ok := err.(*processError); ok {
		perror.header = ""
	}
	perror := newProcessError(message, processName)
	perror.wrapped = err
	return perror
}

// Wrapf wraps the error with the formatted message and arguments.
func Wrapf(err error, processName, message string, arguments ...interface{}) error {
	if perror, ok := err.(*processError); ok {
		perror.header = ""
	}
	perror := newProcessError(fmt.Sprintf(message, arguments...), processName)
	perror.wrapped = err
	return perror
}
