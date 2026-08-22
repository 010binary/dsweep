// Package buildinfo exposes the version metadata stamped into the binary at
// link time.
package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Overwritten via -ldflags -X (see the Makefile). The defaults are what a
// plain `go build` produces.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String renders a one-line version banner.
//
// When Version is still the default, it falls back to the module metadata Go
// embeds automatically, so `go install`-ed builds — which carry no ldflags —
// still report a real version and revision.
func String() string {
	version, commit, date := Version, Commit, Date

	if version == "dev" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if v := bi.Main.Version; v != "" && v != "(devel)" {
				version = v
			}
			for _, s := range bi.Settings {
				switch s.Key {
				case "vcs.revision":
					commit = s.Value
				case "vcs.time":
					date = s.Value
				}
			}
		}
	}

	return fmt.Sprintf("dsweep %s (commit %s, built %s, %s/%s, %s)",
		version, commit, date, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
