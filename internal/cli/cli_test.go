package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/010binary/dsweep/internal/exit"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantCode   exit.Code
		wantStdout string // substring
	}{
		{"no args prints usage", nil, exit.OK, "Usage:"},
		{"help", []string{"help"}, exit.OK, "Usage:"},
		{"short help flag", []string{"-h"}, exit.OK, "Usage:"},
		{"long help flag", []string{"--help"}, exit.OK, "Usage:"},
		{"version", []string{"version"}, exit.OK, "dsweep "},
		{"version flag", []string{"--version"}, exit.OK, "dsweep "},
		{"unknown command", []string{"frobnicate"}, exit.Usage, ""},
		{"unimplemented command is a usage error", []string{"scan"}, exit.Usage, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			err := Run(context.Background(), tt.args, &stdout, &stderr)

			if got := exit.FromError(err); got != tt.wantCode {
				t.Errorf("exit code = %d (%v), want %d", got, err, tt.wantCode)
			}
			if !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Errorf("stdout = %q, want it to contain %q", stdout.String(), tt.wantStdout)
			}
			// Run must stay silent on stderr; main owns diagnostics.
			if stderr.Len() != 0 {
				t.Errorf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestRunHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var stdout, stderr bytes.Buffer
	err := Run(ctx, []string{"version"}, &stdout, &stderr)

	if got := exit.FromError(err); got != exit.Failure {
		t.Errorf("exit code = %d, want %d for a canceled context", got, exit.Failure)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want no output after cancellation", stdout.String())
	}
}

// The usage text documents the exit codes; keep it in sync with the exit
// package so the two cannot drift.
func TestUsageDocumentsEveryExitCode(t *testing.T) {
	t.Parallel()

	for _, code := range []exit.Code{exit.OK, exit.Failure, exit.Usage, exit.Findings} {
		if !strings.Contains(usage, string(rune('0'+int(code)))) {
			t.Errorf("usage text does not mention exit code %d", int(code))
		}
	}
}

// errWriter fails every write, standing in for a closed pipe or a full disk.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("no space left on device") }

func TestRunReportsWriteFailure(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	err := Run(context.Background(), []string{"help"}, errWriter{}, &stderr)

	if got := exit.FromError(err); got != exit.Failure {
		t.Errorf("exit code = %d, want %d when stdout is unwritable", got, exit.Failure)
	}
	if !strings.Contains(err.Error(), "no space left on device") {
		t.Errorf("error = %q, want it to wrap the write failure", err)
	}
}
