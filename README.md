# dsweep

Find and reclaim wasted disk space — safely.

`dsweep` walks a directory tree, identifies reclaimable junk (dependency
directories, build artifacts, stale caches), and tells you what it found before
touching anything. Deletion is opt-in, reversible, and logged.

---

## Status: pre-alpha

**Not usable yet.** The scaffolding is in place — command layer, exit-code
contract, build and test tooling — but the scanner and sweeper are not
implemented. Today the binary answers `help` and `version` and nothing else.

Everything under [Usage](#usage) describes the intended interface, not shipped
behavior. Follow the [roadmap](#roadmap) for progress.

---

## Why

Disk-cleanup tools tend to make the same mistake: they interleave *deciding*
what to remove with *removing* it. Dry-run then becomes a flag checked in a
dozen places, and it drifts out of sync with the real code path. That is how a
cleanup tool eats a directory you cared about.

`dsweep` splits the two halves and never lets them merge:

- **Scan** is pure and read-only. It walks a tree, applies rules, and produces
  a *plan* — a serializable list of targets with sizes and reasons. It cannot
  delete anything, because it has no code to do so.
- **Sweep** consumes a plan and executes it. It is the only part of the
  codebase permitted to unlink a file.

Dry-run is therefore not a flag — it is simply declining to run the second
half. Plans are reviewable artifacts you can inspect, diff, or pipe between
machines.

## Install

Releases are not published yet. Until then, build from source:

```sh
git clone https://github.com/010binary/dsweep
cd dsweep
make build          # -> bin/dsweep
```

Or install straight into `GOBIN`:

```sh
go install github.com/010binary/dsweep/cmd/dsweep@latest
```

Requires Go 1.26 or newer. `dsweep` has no runtime dependencies.

## Usage

> Planned interface. Only `help` and `version` work today.

```sh
dsweep scan ~/code                 # report what could be reclaimed
dsweep scan ~/code --json          # same, machine-readable
dsweep scan ~/code -o plan.jsonl   # save a plan for review

dsweep sweep plan.jsonl            # dry run: print what would happen
dsweep sweep plan.jsonl --apply    # move targets to the trash
dsweep sweep plan.jsonl --apply --purge   # unlink permanently

dsweep undo                        # restore the last sweep
```

### Safety model

Every one of these is a hard rule, not a default you can drift past:

- **Dry-run is the default.** Mutating the filesystem requires an explicit
  `--apply`.
- **Trash first.** Targets move to the platform trash; `--purge` is needed to
  genuinely unlink.
- **Reversible.** A manifest is written *before* removal, so `dsweep undo`
  works and a crash mid-sweep is recoverable.
- **Confined traversal.** Walks are rooted with `os.Root`, so a symlink cannot
  escape the directory you named.
- **Symlinks are never followed on delete.** The link is removed, never its
  target.
- **Protected paths.** `/`, `$HOME`, and volume roots are refused regardless of
  flags. `.git`, `.ssh`, keychains, and `.env` are on a denylist users may
  extend but not shrink.
- **No silent filesystem crossings.**

### Exit codes

Stable across releases — scripts may rely on them.

| Code  | Meaning                                  |
| ----- | ---------------------------------------- |
| `0` | Success, nothing to do                   |
| `1` | Runtime error (I/O, permissions, cancel) |
| `2` | Usage error (bad flag or argument)       |
| `3` | Success, reclaimable files found         |

Note that `3` is a **success**, not an error: it reports a finding and prints
nothing to stderr. This lets you branch without parsing output:

```sh
dsweep scan ~/code
case $? in
  0) echo "already clean" ;;
  3) echo "found something to reclaim" ;;
  *) echo "scan failed" >&2; exit 1 ;;
esac
```

## Development

```sh
make help        # list every target
make check       # fmt + vet + test -race, run this before pushing
make test        # tests with the race detector
make cover       # HTML coverage report
make build-all   # cross-compile to dist/ for all supported platforms
```

### Layout

```
cmd/dsweep/        thin entry point: signals, exit codes, nothing else
internal/cli/      command layer — Run(ctx, args, stdout, stderr) error
internal/exit/     exit-code contract and the error type that carries it
internal/buildinfo version metadata, injected at link time
```

Two rules govern output:

1. **Only `main` prints failures.** `cli.Run` returns an error carrying an exit
   code and writes nothing to stderr, so the message and the status can never
   disagree. Enforced today.
2. **Only `internal/report` will write to stdout.** Structured output has to be
   designed in, not retrofitted onto scattered `fmt.Println` calls. The package
   arrives with the reporters in phase 3; until then `internal/cli` prints its
   own help text.

## Roadmap

| Phase | Scope                                                    | Status |
| ----- | -------------------------------------------------------- | ------ |
| 0     | Build, exit-code contract, repo hygiene                  | done   |
| 1     | Cobra command tree, config layering,`slog`, `--json` | next   |
| 2     | Concurrent scanner, on-disk sizes, hardlink dedup        | —     |
| 3     | Rule engine and reporters                                | —     |
| 4     | Sweeper and the full safety layer                        | —     |
| 5     | `testscript` end-to-end suite                          | —     |
| 6     | CI matrix,`golangci-lint`, GoReleaser                  | —     |
| 7     | Completions, man pages, presets                          | —     |

## License

MIT — see [LICENSE](LICENSE).
