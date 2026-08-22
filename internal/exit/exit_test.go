package exit

import (
	"errors"
	"fmt"
	"testing"
)

func TestFromError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want Code
	}{
		{"nil is success", nil, OK},
		{"uncoded error is a failure", errors.New("boom"), Failure},
		{"usage helper", Usagef("bad flag %q", "-x"), Usage},
		{"failure helper", Failuref("read: %v", errors.New("eof")), Failure},
		{"findings helper", Found(), Findings},
		{"code survives wrapping", fmt.Errorf("scan: %w", Usagef("bad flag")), Usage},
		{"innermost coded error wins", fmt.Errorf("outer: %w", Found()), Findings},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := FromError(tt.err); got != tt.want {
				t.Errorf("FromError(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

func TestDiagnostic(t *testing.T) {
	t.Parallel()

	// Only genuine failures should produce stderr output. Findings is a
	// successful outcome that merely sets a non-zero status.
	for code, want := range map[Code]bool{
		OK:       false,
		Failure:  true,
		Usage:    true,
		Findings: false,
	} {
		if got := code.Diagnostic(); got != want {
			t.Errorf("Code(%d).Diagnostic() = %v, want %v", int(code), got, want)
		}
	}
}

func TestErrorMessage(t *testing.T) {
	t.Parallel()

	cause := errors.New("permission denied")
	coded := &Error{Code: Failure, Err: cause}

	if got := coded.Error(); got != "permission denied" {
		t.Errorf("Error() = %q, want the cause's message", got)
	}
	if !errors.Is(coded, cause) {
		t.Error("errors.Is could not reach the wrapped cause")
	}

	// A code with no cause still needs a usable message.
	if got := Found().Error(); got == "" {
		t.Error("Found().Error() is empty; want a description")
	}
}

// The numeric values are a public contract: scripts branch on them.
func TestCodeValuesAreStable(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		code Code
		want int
	}{{OK, 0}, {Failure, 1}, {Usage, 2}, {Findings, 3}} {
		if int(tt.code) != tt.want {
			t.Errorf("%v = %d, want %d", tt.code, int(tt.code), tt.want)
		}
	}
}
