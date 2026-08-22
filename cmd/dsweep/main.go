// Command dsweep finds and reclaims wasted disk space.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/010binary/dsweep/internal/cli"
	"github.com/010binary/dsweep/internal/exit"
)

func main() {
	// os.Exit skips deferred calls, so all real work happens in run and
	// main stays a single statement.
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr)

	code := exit.FromError(err)
	if code.Diagnostic() {
		fmt.Fprintf(os.Stderr, "dsweep: %v\n", err)
	}
	return int(code)
}
