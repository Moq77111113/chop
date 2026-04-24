# AGENTS.md

> Guidelines for AI coding agents working on the chop codebase.

## Project Overview

**chop** is a local streaming / network impairment testbench. Spin up synthetic cameras (or wire real ones), insert proxy "links" between them and consumers, and degrade the network in real time via a dashboard (loss, latency, jitter, bandwidth, kill stream). Built to validate video pipelines (compositors, encoders, downstream services) under realistic field conditions before they hit production.

### Tech Stack

- **Backend**: Go (latest stable), `github.com/bluenviron/gortsplib/v4` for RTSP, `gopkg.in/yaml.v3` for scenario parsing, `gorilla/websocket` for the dashboard live channel
- **Frontend**: SolidJS + Vite + TypeScript (latest), `embed.FS` to ship the SPA inside the Go binary
- **IPC**: stdio JSON-RPC (ndjson) between supervisor and block child processes
- **CI**: GitHub Actions, `golangci-lint`, `make smoke`

## Architecture

Single Go module, single binary `chop` with two modes : `chop run scenario.yml` (supervisor + dashboard) and `chop block <type>` (one block child, called by the supervisor itself). Each block runs as its own process — crash isolation, language portability later, supervisor talks to all of them through stdio JSON-RPC.

```
chop/
├── cmd/chop/                    # CLI dispatch (run | lint | block)
├── block/                       # *PUBLIC* framework surface
│   ├── block.go                   # Block interface + Info / Snapshot / Config
│   ├── runtime.go                 # RunBlock — wires stdio JSON-RPC to Block
│   ├── emit.go                    # Emit(ctx, kind, payload) for events
│   └── methods.go                 # MethodInfo / MethodSnapshot / MethodApply / MethodAction
├── internal/blocks/             # built-in block types (NOT importable externally)
│   ├── source/                    # synthetic RTSP source (h264 file replay → server)
│   ├── link/                      # RTSP proxy with impairments (loss / lat / jitter / bw)
│   └── process/                   # wrap any external binary (mediamtx, gst, ffmpeg, ...)
├── internal/transport/          # ndjson JSON-RPC framing (hidden by RunBlock)
├── internal/scenario/           # YAML loader + validator (zero side effect)
├── internal/supervisor/         # spawn, lifecycle, registry
│   └── api/                       # HTTP routes + WebSocket
├── internal/dashboard/          # embed.FS for the built SPA
├── web/                         # SolidJS + Vite sources
├── examples/                    # canonical scenarios (smoke, panoramic, ...)
├── testdata/                    # fixtures (h264 file etc., gitignored)
└── scripts/                     # smoke.sh and similar
```

`block/` is the only **public** package — external devs writing their own blocks (in Go, or in any language via the wire protocol) import `github.com/moq77111113/chop/block`. Everything under `internal/` is private.

### Visibility / dependency rules

| Package                       | Visibility | Responsible for                              | Never touches                       |
|-------------------------------|------------|----------------------------------------------|-------------------------------------|
| `block`                       | public     | Block interface + runtime stdio              | gortsplib, scenario, HTTP           |
| `internal/blocks/source`      | internal   | synthetic / passthrough RTSP source           | supervisor, scenario, dashboard     |
| `internal/blocks/link`        | internal   | RTSP proxy + per-packet impairments           | scenario, dashboard, autres blocs   |
| `internal/blocks/process`     | internal   | exec.Cmd wrapper + ready check                | tout le reste                       |
| `internal/scenario`           | internal   | YAML parse + validate + lint                  | exec, sockets, gortsplib            |
| `internal/supervisor`         | internal   | DAG, spawn, lifecycle, registry               | logique métier des blocs            |
| `internal/supervisor/api`     | internal   | HTTP routes + WebSocket                       | logique métier des blocs            |
| `internal/transport`          | internal   | ndjson JSON-RPC framing                       | tout le reste                       |
| `internal/dashboard`          | internal   | embed.FS de la SPA                            | rien d'autre                        |
| `web/`                        | sources    | SPA                                            | tout backend                        |
| `cmd/chop`                    | internal   | CLI dispatch                                   | aucune logique métier               |

A dependency that violates this table is an architecture bug. Rule of thumb : the table goes top → bottom in import direction. Arrows go **down**, never up.

## Design Principles

### Block contract is the only public API

`block.Block` (5 methods) + `block.RunBlock` + `block.Emit` + the `block.Method*` constants. That's the entire surface external devs see. No frame internals leak out — no transport types, no supervisor types, no HTTP routes.

### Process-per-block

Each block runs in its own OS process, supervisor spawns it via `exec.Cmd`. Crash isolation, native cross-language plug-ins later (Rust block = another binary speaking the same stdio protocol). The cost (~10MB per Go process) is negligible vs the failure mode of a goroutine panic taking down the whole testbench.

### stdio JSON-RPC over HTTP for IPC

Supervisor ↔ blocks talk via ndjson on the child's stdin/stdout. No port allocation per block. EOF on the pipe = block dead, instant detection. stderr stays for free-form logs (captured by the supervisor, surfaced in the dashboard). The browser ↔ supervisor link is HTTP/WS (no choice — it's a browser).

### Pure logic, isolated effects

The decision-making code (impairment math, DAG topo-sort, YAML validate) is pure — no sockets, no spawning, no time-based randomness sneaked in. The effectful code (gortsplib server, exec.Cmd, HTTP handlers) calls the pure code. This keeps unit tests fast and meaningful.

### No god types

Each block type defines its own `Controls` struct (`LinkControls{Loss, LatencyMs, ...}`), its own `Snapshot.Stats` shape (`LinkSnapshot{...}`). Type-safe inside the block, type-erased on the wire (`json.RawMessage`). The `block` package never grows a `MasterConfig` that knows about every block type.

### Owner-our-vocabulary

User-facing names belong to chop, not to the libraries we use. The YAML says `pattern: bouncing-ball`, the framework maps it internally to `videotestsrc pattern=ball`. If we swap GStreamer for libav tomorrow, the YAML doesn't change.

## Build & Dev Commands

```bash
make build                   # build front + embed dist + go build
make test                    # go test -race ./...
make smoke                   # build + run scripts/smoke.sh end-to-end
make lint                    # golangci-lint
go test -race ./block/...    # test one package quickly
```

Per-package iteration is preferred during dev (fast feedback). `make smoke` is the truth gate before merging.

## Code Conventions

### Files & functions

- **One file = one responsibility.** Target < 100 LOC per file ; split when it grows past that for any reason other than a single coherent block.
- **Functions read like sentences.** Extract named helpers (`lostToDice`, `latencyWithJitter`, `dropped`, `delayed`) instead of inlining logic.
- **Guard clauses early.** No nested ifs.
- **Helpers below their caller**, in the order they're encountered.
- **Avoid "and" in function names.** `parseAndValidate` is two functions wearing one mask.

### Naming

- **No magic numbers, no magic strings.** Every literal that has a name should have a name. RPC verbs (`MethodSnapshot = "snapshot"`), exit codes (`exitBlockRunErr = 1`), buffer sizes (`scanInitialBuf = 64 * 1024`), repeated empty payloads (`emptyAck = json.RawMessage("{}")`).
- **Exported = part of the public contract.** Capitalized identifiers in `block/` are forever-stable. Add carefully.
- **Type names describe what, helper names describe what they do.** `Controls`, `Snapshot`, `Decision` (nouns) ; `Decide`, `Apply`, `Emit` (verbs).

### No comments

Code is self-explanatory. If a comment is needed to understand WHAT, refactor the code. Comments justify WHY only when non-obvious : a hidden invariant, a workaround for an upstream bug, a perf-critical choice. Doc comments on exported APIs (`block.Block`, `block.RunBlock`) are part of the public surface and stay tight.

### Types over flags

Discriminated configs use a `kind:` field that maps to a typed shape. No `if cfg.IsRTSP { ... } else if cfg.IsSynthetic { ... }` chains.

### KISS / YAGNI

- The MVP ships only what the next test needs.
- No future-proofing for hypotheticals. The code already flags clear extension points (e.g. backend `proxy | netem` for the link impairment) — ship one, design the seam, don't build the second until someone asks.
- DRY but not at the cost of readability — three similar lines beat one premature abstraction.

## Testing

- **Tests test intent, not coverage.** Each test is named after the behavior it verifies (`TestLink_LossMatchesControl`, not `TestLink_RandFloat`). The question for every test : *if I delete this, what user-visible thing breaks ?* If the answer is internal mechanics, the test is wrong.
- **Pure logic = unit tests in the same package** (`impair_test.go` next to `impair.go`). No mocks, no sockets, no time skew via real clock.
- **Integration tests** spin up real components (real gortsplib server, real proxy) and assert observable behavior. Located under `internal/integration/` for cross-package flows, or alongside the package for narrow cases.
- **Smoke test** = `scripts/smoke.sh`, runs end-to-end in CI in < 30s.
- **Race detector always.** `go test -race ./...` is the default. CI fails otherwise.
- **Tests don't drive coverage targets.** ~80% on pure logic is plenty. Glue code (HTTP shims, JSON-RPC) is covered by integration, not unit.

## Git

- **Conventional commits** : `feat(scope): ...`, `fix(scope): ...`, `refactor: ...`, `chore: ...`, `docs: ...`, `test: ...`. Scope is the package or area touched.
- **Short message only.** No body unless absolutely necessary. No co-author signature, no AI mention.
- **First commit on a fresh repo** : `first commit` (humble, formal pattern starts at commit #2).
- **Atomic commits.** One concern per commit. Refactors and feature work go in separate commits.
- **Never `git commit --amend`.** Always create a new commit.
- **Never `git config`, `gh repo create`, `git push`** in agent code. The maintainer does these manually.
- **Branch pattern** (when branching is used) : `{type}/{short-description}`.

## Frontend (web/)

The dashboard is dev tooling — clarity over polish. SolidJS + Vite + TS. No design system to enforce here ; one design intent : pipeline visible at a glance, controls reactive without lag.

- **Solid signals everywhere** for live data. Avoid React-style re-render thinking.
- **WebSocket for live state**, fetch for one-shots. Debounce slider-driven `PATCH /api/blocks/:id/controls` (50ms is plenty).
- **One file per component.** Same < 100 LOC target.
- **No CSS framework imposed.** Inline styles or a tiny shared `styles.css` are both fine while the surface is small. Reassess if it grows.
- **Built artifacts** go to `web/dist/` then are copied to `internal/dashboard/dist/` for `embed.FS`. Don't commit either.

## Verification before claiming done

1. `go test -race ./...` passes (or the targeted package + downstream)
2. `go vet ./...` clean
3. `golangci-lint run` clean (or new warnings only on lines you touched, with reason)
4. `git status` clean, no untracked files left behind
5. New diagnostic from your editor / LSP is acted on, not ignored

For features touching the runtime end-to-end : `make smoke` green locally before commit.
