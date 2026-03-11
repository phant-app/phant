# Phant Architecture Overview

Date: 2026-03-11
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

### `setup.PHPManagerSnapshot`

Defined in [internal/setup/php_manager_types.go](../../internal/setup/php_manager_types.go).

Linux-first snapshot for PHP Manager UI:

- active + installed/available PHP versions
- selected php.ini settings (`upload_max_filesize`, `post_max_size`, `memory_limit`, `max_execution_time`)
- extension inventory and enabled state
- warnings/errors for unsupported platforms or command failures

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
- powers PHP manager operations for Linux:
	- discover versions/settings/extensions
	- install and switch PHP versions
	- apply managed php.ini settings to CLI + discovered FPM targets
	- toggle extensions and restart discovered PHP-FPM services
- returns diagnostics/install results for UI operations

### `main` package (`app.go`, `main.go`)

Responsibility: app orchestration and Wails bridge.

- startup/shutdown lifecycle hooks
- starts/stops collector
- bridges collector events to frontend via Wails runtime channel
- exposes frontend-callable methods through dedicated services (`DumpService`, `SetupService`, `PHPService`)

### `internal/app/phpmanager`

Responsibility: PHP manager use-cases at application layer.

- orchestrates snapshot/action flows for PHP versions, settings, and extensions
- currently delegates to setup internals while migration is in progress
- gives a stable seam to migrate Linux/macOS/Windows providers without changing Wails service contracts

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

### 6) Service boundaries by capability

Decision:

- `SetupService` is environment/setup focused (diagnostics, CLI hook, Valet verification/remediation)
- `PHPService` is PHP manager focused (versions, switch, settings, extensions)

Why:

- avoids an all-purpose setup bag
- keeps frontend bindings grouped by domain intent
- reduces coupling as PHP manager evolves independently

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
- Linux-first PHP manager backend and frontend wiring (versions + settings + extension toggles)
- Wails service split: dedicated `PHPService` plus app-layer `internal/app/phpmanager` seam

## Remaining work (next stage)

Primary next milestones:

1. Harden PHP manager safety UX (dry-run/preview + richer partial-failure reporting)
2. macOS/Windows providers for PHP manager and privileged automation
3. Improve remediation safety UX (preview diff + per-service selective apply)
4. Dump transport fallback mode (Symfony dump server adapter)
5. Persistence and retention controls

## File map (quick navigation)

- App orchestration: [app.go](../../app.go), [main.go](../../main.go)
- Wails services: [internal/services/dump_service.go](../../internal/services/dump_service.go), [internal/services/setup_service.go](../../internal/services/setup_service.go), [internal/services/php_service.go](../../internal/services/php_service.go)
- Dump domain: [internal/dump/types.go](../../internal/dump/types.go), [internal/dump/decoder.go](../../internal/dump/decoder.go)
- Collector: [internal/collector/server.go](../../internal/collector/server.go), [internal/collector/buffer.go](../../internal/collector/buffer.go), [internal/collector/path.go](../../internal/collector/path.go)
- Setup diagnostics: [internal/setup/diagnostics.go](../../internal/setup/diagnostics.go)
- Setup hook installer: [internal/setup/hook_installer.go](../../internal/setup/hook_installer.go)
- Valet verifier: [internal/setup/valet_linux.go](../../internal/setup/valet_linux.go)
- PHP manager: [internal/setup/php_manager.go](../../internal/setup/php_manager.go), [internal/setup/php_manager_linux.go](../../internal/setup/php_manager_linux.go), [internal/setup/php_manager_types.go](../../internal/setup/php_manager_types.go)
- PHP manager app layer: [internal/app/phpmanager/service.go](../../internal/app/phpmanager/service.go)
- Frontend live UI: [frontend/src/App.tsx](../../frontend/src/App.tsx)
- PHP manager UI: [frontend/src/pages/PhpManagerPage.tsx](../../frontend/src/pages/PhpManagerPage.tsx)
- Schema spec: [docs/specs/dump-event-schema.md](../specs/dump-event-schema.md)