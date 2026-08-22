// Package cli is dsweep's command layer.
//
// Run is the only entry point: cmd/dsweep does nothing but call it and turn
// its error into a process exit code. Keeping the whole command tree behind
// one function with injected writers means every command is testable without
// spawning a process or touching the real os.Stdout.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/010binary/dsweep/internal/buildinfo"
	"github.com/010binary/dsweep/internal/exit"
)

const usage = `dsweep — find and reclaim wasted disk space.

Usage:
  dsweep <command> [flags]

Commands:
  version      Print version information
  help         Show this help

Not implemented yet: scan, sweep, undo.

Exit codes:
  0  success, nothing to do
  1  runtime error
  2  usage error
  3  success, reclaimable files found
`

// Run executes dsweep and returns an error carrying an [exit.Code].
//
// stdout receives program output; stderr is reserved for diagnostics and is
// currently unused — Run never prints its own failures, so that main owns the
// single place where an error becomes user-visible text.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if err := ctx.Err(); err != nil {
		return &exit.Error{Code: exit.Failure, Err: err}
	}

	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return nil

	case "version", "--version":
		fmt.Fprintln(stdout, buildinfo.String())
		return nil

	default:
		return exit.Usagef("unknown command %q; run 'dsweep help' for usage", args[0])
	}
}
