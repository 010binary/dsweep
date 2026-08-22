// Package exit defines dsweep's process exit codes and the error type that
// carries them out of the command layer.
//
// These codes are part of dsweep's public contract — scripts branch on them,
// so they must stay stable across releases.
package exit

import (
	"errors"
	"fmt"
)

// Code is a process exit status.
type Code int

const (
	// OK means the command completed with nothing to act on.
	OK Code = 0

	// Failure means the command hit a runtime error: I/O, permissions,
	// cancellation, anything unexpected.
	Failure Code = 1

	// Usage means the command was invoked incorrectly: bad flag, bad
	// argument, unknown subcommand.
	Usage Code = 2

	// Findings means the command succeeded and found reclaimable files.
	// It is a success, not an error: `dsweep scan` reports it so callers
	// can branch without parsing output.
	Findings Code = 3
)

func (c Code) String() string {
	switch c {
	case OK:
		return "ok"
	case Failure:
		return "error"
	case Usage:
		return "usage error"
	case Findings:
		return "reclaimable files found"
	default:
		return fmt.Sprintf("unknown code %d", int(c))
	}
}

// Diagnostic reports whether the code represents a failure worth printing a
// message for. OK and Findings are both successful outcomes and stay quiet.
func (c Code) Diagnostic() bool {
	return c == Failure || c == Usage
}

// Error pairs an exit Code with an optional cause. The command layer returns
// it so that a single place — main — decides both the process status and
// whether the user sees a message.
type Error struct {
	Code Code
	Err  error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code.String()
}

func (e *Error) Unwrap() error { return e.Err }

// Failuref returns an error that exits with [Failure].
func Failuref(format string, a ...any) error {
	return &Error{Code: Failure, Err: fmt.Errorf(format, a...)}
}

// Usagef returns an error that exits with [Usage].
func Usagef(format string, a ...any) error {
	return &Error{Code: Usage, Err: fmt.Errorf(format, a...)}
}

// Found returns an error that exits with [Findings]. It reports no failure and
// produces no stderr output; it exists only to set the process status.
func Found() error {
	return &Error{Code: Findings}
}

// FromError maps an error from the command layer onto a process exit code. A
// nil error is [OK]; an error carrying no explicit code is a [Failure].
func FromError(err error) Code {
	if err == nil {
		return OK
	}
	var coded *Error
	if errors.As(err, &coded) {
		return coded.Code
	}
	return Failure
}
