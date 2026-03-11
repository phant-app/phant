# Phant Architecture Migration Implementation Plan

Date: 2026-03-11
Status: Approved to execute incrementally

## Goal in simple terms

Restructure Phant so each capability has a clear home.

- `SetupService` should only handle environment/setup concerns.
- `PHPService` should own PHP manager concerns.
- Core PHP manager models should live in a domain package.
- OS-specific command and filesystem behavior should move out of mixed setup files into infrastructure adapters.

This keeps Wails3 bindings clean and makes future features easier to add without creating another "bag" package.

## Plan strategy

1. Migrate in small, behavior-preserving slices.
2. Keep temporary compatibility layers during the transition.
3. Regenerate Wails bindings after each service signature move.
4. Verify Go tests and frontend build at every slice.
5. Remove compatibility shims only after frontend/backend are stable.

## Step-by-step tasks

### Phase 1: Contract and boundary foundation (safe, no behavior change)

1. Introduce `internal/domain/phpmanager/types.go`.
- Purpose: Make PHP manager models belong to a domain package, not setup.
- Files:
  - create `internal/domain/phpmanager/types.go`
  - update `internal/setup/php_manager_types.go` to type aliases for compatibility
- Steps:
  1. Move structs (`PHPVersion`, `PHPIniSettings`, `PHPExtension`, snapshot/action request/result types) to domain package.
  2. Replace setup structs with aliases to the new domain types.
- Verification:
  - `go test ./...`
  - `wails3 generate bindings -ts && npm run -s build`
- Rollback hint:
  - Revert alias file to concrete setup structs if binding generation fails.

2. Move service signatures to domain types.
- Purpose: Ensure application and Wails service layers speak domain types directly.
- Files:
  - `internal/app/phpmanager/service.go`
  - `internal/services/php_service.go`
- Steps:
  1. Update method signatures to return/accept domain models.
  2. Keep setup compatibility wrappers working by relying on alias equivalence.
- Verification:
  - Regenerate bindings and confirm `frontend/src/pages/PhpManagerPage.tsx` still compiles.
- Rollback hint:
  - Temporarily keep service signatures in setup types while domain types stabilize.

### Phase 2: Use-case extraction (application layer owns orchestration)

3. Move orchestration from setup to app layer.
- Purpose: `internal/app/phpmanager` should orchestrate snapshot + action flow.
- Files:
  - create `internal/app/phpmanager/usecases.go`
  - update `internal/setup/php_manager.go` to delegation wrappers
- Steps:
  1. Copy orchestration logic into app layer.
  2. Keep setup functions as compatibility wrappers that call app layer.
- Verification:
  - Existing behavior for snapshot/install/switch/settings/extensions remains unchanged.
- Rollback hint:
  - Point wrappers back to setup logic if parity regressions appear.

### Phase 3: Infrastructure extraction (OS and shell logic)

4. Extract Linux command logic into infrastructure adapters.
- Purpose: Separate "what to do" from "how commands/files are executed."
- Files:
  - create `internal/infra/php/linux/provider.go`
  - create `internal/infra/system/runner.go`
  - gradually shrink `internal/setup/php_manager_linux.go`
- Steps:
  1. Move command execution and parsing helpers to adapter package.
  2. Inject adapter into app use-case service.
  3. Keep setup compatibility wrappers until complete migration.
- Verification:
  - Unit tests for parser and command-path errors still pass.
- Rollback hint:
  - Temporarily bind app layer to legacy setup implementation if adapter injection has gaps.

### Phase 4: Setup package narrowing

5. Keep only environment concerns in setup package.
- Purpose: setup package becomes focused and intentional.
- Files:
  - `internal/setup/diagnostics.go`
  - `internal/setup/hook_installer.go`
  - `internal/setup/valet_linux.go`
  - `internal/setup/valet_sites.go`
- Steps:
  1. Remove migrated PHP manager files from setup package.
  2. Delete compatibility wrappers after all callers are migrated.
- Verification:
  - Wails bindings no longer expose PHP manager models under setup package.
- Rollback hint:
  - Reintroduce thin wrappers for one release window if downstream clients still depend on setup models.

### Phase 5: Service/API cleanup and documentation

6. Final service shape and docs.
- Purpose: Align architecture docs with code reality.
- Files:
  - `internal/services/setup_service.go`
  - `internal/services/php_service.go`
  - `docs/architecture/overview.md`
  - `README.md`
- Steps:
  1. Confirm SetupService has only setup methods.
  2. Confirm PHPService has all PHP manager methods.
  3. Update architecture diagrams and package responsibility docs.
- Verification:
  - Full Go + frontend build.
- Rollback hint:
  - Restore previous binding generation and service signatures if doc mismatch exposes unresolved code references.

## Workflow and why

This order is safest because:

1. Types and service contracts move first with aliases to avoid breakage.
2. Behavior moves second after contracts are stable.
3. Infrastructure moves third so command/file complexity is isolated last.
4. Cleanup is done only when tests and bindings are consistently green.

## Verification checkpoints

1. After each phase, run:
- `go test ./...`
- `wails3 generate bindings -ts`
- `cd frontend && npm run -s build`

2. Manual checks:
- PHP page still loads versions/settings/extensions.
- Install/switch/settings/extensions actions still call the backend correctly.
- Settings page diagnostics and valet panels remain unaffected.

## Plan file path

`docs/plans/2026-03-11-architecture-migration-implementation-plan.md`

## Review before coding

The plan is ready and saved. The next execution starts with Phase 1 (domain types + compatibility aliases + service signatures) as a behavior-preserving migration slice.
