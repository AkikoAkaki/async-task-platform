# Development Log (AI Contributors)

This log captures working agreements, version semantics, and milestone context so AI or human collaborators share the same mental model.

## Project Overview

**Async Task Platform** is a distributed asynchronous task scheduling system designed to handle:

1. **Delayed Execution**: Execute tasks at a precise future time (e.g., cancel unpaid orders after 30 minutes)
2. **Periodic Scheduling**: Execute tasks on a recurring schedule (e.g., daily reports) — *planned*
3. **Workflow Orchestration**: Chain multiple tasks with dependencies (e.g., user onboarding flow) — *planned*

The current implementation (MVP) focuses on delayed execution with Redis-backed storage and gRPC APIs.

## Working Agreements

### Branching & Reviews
- Follow Git Flow (see ADR-002): `develop` is integration, `main` mirrors production, `feature/<slug>` branches merge via PRs.
- Hotfixes fork from `main` and are back-merged into `develop` after release.
- Every PR references at least one ADR or creates a draft ADR when making non-trivial architectural/process choices.

### Versioning & Tags
- SemVer (`MAJOR.MINOR.PATCH`). Breaking API/storage changes bump `MAJOR`, backward-compatible feature sets bump `MINOR`, and fixes/ops tweaks bump `PATCH`.
- Tags are annotated in the format `vX.Y.Z` and applied on `main` after the release branch is merged.
- Next milestone release is expected to become `v0.2.0` once Delete API and Retrieve gRPC endpoint are complete.

### Changelog & Tooling
- Maintain `CHANGELOG.md` manually with the Keep a Changelog format. Use `git-chglog --next-tag vX.Y.Z` to hydrate entries, then reconcile wording before committing.
- Keep a `.chglog/config.yml` that encodes section ordering so generators stay deterministic.
- Run `make fmt && make lint && make test` before opening a PR to ensure generated code stays tidy.

### Documentation Rituals
- Update ADRs (`docs/adr/`) for any decision with architectural or operational impact. Start as `Draft`, promote to `Accepted` post-merge, link replacing ADRs when superseding.
- Record summary bullets of notable work here (see Change Journal) so AI agents can pick up context quickly.
- Mirror release-impacting updates inside `docs/ARCHITECTURE.md` or `docs/API.md` as needed.

## Release Workflow & CI Integration

1. **Branch cut:** When a milestone on `develop` hardens, create `release/vX.Y.Z` (use SemVer). Push immediately so CI attaches provenance.
2. **Docs first:** Update `CHANGELOG.md` with `[X.Y.Z] - YYYY-MM-DD` and summarize scope; add a matching note under `Milestones` in this file.
3. **Automated checks:** Every push to `release/*` runs the `release-readiness` job (see `.github/workflows/ci.yaml`). It validates branch naming, ensures both `CHANGELOG.md` and `DEVELOPMENT_LOG.md` mention the release, installs `git-chglog`, and uploads a preview generated via `git-chglog --next-tag vX.Y.Z` for reviewers.
4. **Stabilization:** Only release-blocking fixes land on the release branch. All other work stays on `develop`.
5. **Cut & tag:** After CI passes, merge `release/vX.Y.Z` into `main`, tag `vX.Y.Z` with release notes, then merge back into `develop`.
6. **Post-release:** Archive the generated release notes artifact (kept 7 days) into the GitHub Release description along with links to relevant ADRs.

## Decision Index

| ADR | Title | Status |
|-----|-------|--------|
| [ADR-001](docs/adr/001-architecture-and-storage.md) | Redis-based MVP with gRPC surface | Accepted |
| [ADR-002](docs/adr/002-gitflow-and-versioning.md) | Git Flow adoption and SemVer policy | Accepted |

## Roadmap

### Phase 1: Core Completion (Current Focus)
- [ ] Implement `Delete` API for task cancellation
- [ ] Implement `Retrieve` gRPC endpoint
- [ ] Add idempotency key support for Enqueue
- [ ] Task priority support (encoded in ZSet score)

### Phase 2: Distributed Scheduling
- [ ] Cron expression parsing (`robfig/cron`)
- [ ] Periodic task model extension
- [ ] Leader election (Redis or etcd based)
- [ ] Topic-based queue sharding

### Phase 3: Production Readiness
- [ ] Prometheus metrics endpoint
- [ ] OpenTelemetry tracing integration
- [ ] Redis Sentinel/Cluster support
- [ ] Chaos testing with toxiproxy

### Phase 4: Workflow Engine (Future)
- [ ] DAG-based task dependencies
- [ ] Workflow state persistence
- [ ] Sub-task spawning
- [ ] Visual workflow editor

## Milestones

| Version | Status | Focus |
|---------|--------|-------|
| v0.1.0 | In Progress | Redis-backed delay queue with gRPC API, Watchdog recovery |
| v0.2.0 | Planned | Delete API, Retrieve gRPC, idempotency, metrics foundation |
| v0.3.0 | Planned | Cron scheduling, Leader election |
| v1.0.0 | Future | Production-ready with observability and HA support |

## Change Journal

| Date | Summary |
|------|---------|
| 2026-02-01 | **Major repositioning**: Renamed project vision from "Distributed Delay Queue" to "Async Task Platform". Updated README, ARCHITECTURE.md, API.md, and ADR-001 to reflect new positioning and roadmap. |
| 2026-01-08 | Introduced Git Flow + SemVer governance (ADR-002), drafted CHANGELOG scaffolding for `v0.1.0`, and published this development log for AI contributors. |
| 2026-01-08 | Wired git-chglog config/template, added release-readiness automation to CI, and enumerated HA milestone tasks for `v0.2.0` planning. |
