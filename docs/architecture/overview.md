# Phant Architecture Overview

Date: 2026-03-03
Status: Current implementation baseline

## Why this document exists

This file explains the current backend/frontend architecture in simple terms:

- entities (data structures)
- package responsibilities
- key design decisions
- runtime working flow
- current progress and missing pieces

## High-level architecture

Phant currently works as a local event pipeline:

1. A producer (eventually PHP hook / dump sender) writes one dump event per line (NDJSON) to a Unix socket.
2. `internal/collector` receives and decodes each line.
3. Valid events are kept in a bounded ring buffer.
4. `app.go` bridges collector events to Wails runtime events.
5. React UI loads recent events and listens for live events.

## Entities (core data models)

### `dump.Event`

Defined in [internal/dump/types.go](../../internal/dump/types.go).

Represents one captured dump (from `dump()` or `dd()`), including:

- event identity and schema version
- source context (`http`, `cli`, `worker`, `cron`)
- payload and trace
- host/process metadata

### `CollectorStatus`

Defined in [app.go](../../app.go).

Frontend-facing runtime status:

- running state
- socket path
- last error
- dropped events count

### `setup.Diagnostics`

Defined in [internal/setup/diagnostics.go](../../internal/setup/diagnostics.go).

Environment diagnostics snapshot:

- PHP availability/version
- `php --ini` output
- detected service manager (`systemd`, `service`, `unknown`)
- last diagnostics error

## Package responsibilities

### `internal/dump`

Responsibility: schema contract + strict decoding/validation.

- decode one NDJSON line into `Event`
- validate required fields and types
- validate source-specific rules and schema version

This package does not know about sockets, Wails, or UI.

### `internal/collector`

Responsibility: ingestion runtime and fan-out.

- manages Unix socket server lifecycle
- reads lines from connections
- uses `dump.DecodeNDJSONLine` for parsing
- stores recent events in ring buffer
- broadcasts events to subscribers

This package does not know about React or Wails runtime APIs.

### `internal/setup`

Responsibility: setup diagnostics + hook installation.

- runs local environment checks for setup readiness
- writes Phant prepend script in user config space
- enables CLI hook through conf.d-first strategy
- applies OS-aware privilege strategy for protected paths
	- Linux: attempts `pkexec` for privileged write
	- macOS/Windows: returns manual/admin guidance (no auto-elevation yet)
- verifies Valet Linux / PHP-FPM wiring for `auto_prepend_file` propagation
- applies optional Valet Linux remediation (guarded apply with explicit confirmation)
- returns diagnostics/install results for UI operations

### `main` package (`app.go`, `main.go`)

Responsibility: app orchestration and Wails bridge.

- startup/shutdown lifecycle hooks
- starts/stops collector
- bridges collector events to frontend via Wails runtime channel
- exposes frontend-callable methods (`GetRecentEvents`, `GetCollectorStatus`, `GetSetupDiagnostics`)

## Architectural decisions and rationale

### 1) Use `internal/*` packages

Decision: keep core code under `internal`.

Why:

- enforces module-private boundaries in Go
- prevents accidental external coupling
- keeps package responsibilities explicit

### 2) Use NDJSON over Unix socket

Decision: one JSON event per line over local Unix domain socket.

Why:

- easy stream framing (`\n` = one event)
- low overhead and local-only transport
- simple failure handling per line

Reference schema: [docs/specs/dump-event-schema.md](../specs/dump-event-schema.md)

### 3) Keep a ring buffer

Decision: bounded in-memory retention with dropped-count tracking.

Why:

- protects app from memory spikes during noisy dumps
- keeps latest useful events available for UI
- explicit dropped-count provides visibility under pressure

### 4) Push + pull model to frontend

Decision:

- Pull: `GetRecentEvents(limit)` for initial state.
- Push: Wails runtime event channel (`phant:dump:event`) for live updates.

Why:

- fast initial hydration
- realtime updates without tight polling loops

### 5) Conf.d-first hook install with OS-aware privileges

Decision:

- prefer writing a dedicated `99-phant.ini` in PHP additional-ini scan dir
- patch `php.ini` only when no additional-ini scan dir is available
- on permission failure, follow OS-specific strategy

Why:

- isolates Phant changes from base `php.ini` when possible
- keeps rollback and inspection simple (single dedicated file)
- Linux desktop environments can support in-app elevation via `pkexec`
- preserves safety by returning explicit manual commands when elevation is unavailable

## Working flow (runtime sequence)

1. Wails starts app.
2. `App.startup()` creates and starts `collector.Server`.
3. `App` subscribes to collector and emits each event through runtime channel.
4. Frontend loads initial events + status + diagnostics.
5. Frontend listens to runtime channel and appends incoming events.
6. On shutdown, app stops event bridge, unsubscribes, and stops collector.

## Current progress (implemented)

- dump schema + decoder/validator
- collector socket server + buffer + subscriber fan-out
- lifecycle wiring (startup/shutdown)
- app bridge methods for status and recent events
- runtime event emission to frontend
- live event log UI
- setup diagnostics backend + UI panel
- CLI hook installer with OS-aware privilege strategy (Linux `pkexec` + manual fallback)
- frontend listener lifecycle hardening and duplicate-event protection for runtime stream
- prepend hook stability fix so `dd()` emits a single event (no double emit)
- verified hook rewrite path (`Enable CLI Hook`) updates user prepend script in config dir
- Valet Linux verification report (CLI conf.d + discovered FPM services + recommendations)
- Valet Linux guarded remediation action (writes FPM hook ini + restart attempts with sudo fallback commands)

## Remaining work (next stage)

Primary next milestones:

1. Improve remediation safety UX (preview diff + per-service selective apply)
2. macOS/Windows privileged automation for hook install
3. Dump transport fallback mode (Symfony dump server adapter)
4. Persistence and retention controls
5. Richer UI filters/grouping and better diagnostics UX

## File map (quick navigation)

- App orchestration: [app.go](../../app.go), [main.go](../../main.go)
- Dump domain: [internal/dump/types.go](../../internal/dump/types.go), [internal/dump/decoder.go](../../internal/dump/decoder.go)
- Collector: [internal/collector/server.go](../../internal/collector/server.go), [internal/collector/buffer.go](../../internal/collector/buffer.go), [internal/collector/path.go](../../internal/collector/path.go)
- Setup diagnostics: [internal/setup/diagnostics.go](../../internal/setup/diagnostics.go)
- Setup hook installer: [internal/setup/hook_installer.go](../../internal/setup/hook_installer.go)
- Valet verifier: [internal/setup/valet_linux.go](../../internal/setup/valet_linux.go)
- Frontend live UI: [frontend/src/App.tsx](../../frontend/src/App.tsx)
- Schema spec: [docs/specs/dump-event-schema.md](../specs/dump-event-schema.md)