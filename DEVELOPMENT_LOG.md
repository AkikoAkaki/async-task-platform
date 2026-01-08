# DEVELOPMENT LOG (AI Contributors)

This log captures working agreements, version semantics, and milestone context so AI or human collaborators share the same mental model.

## Purpose
- Provide a single reference for contribution etiquette (Git Flow, SemVer, changelog duties).
- Track lightweight decision context between formal ADRs and commit history.
- Surface the current milestone status and TODO anchors for autonomous agents.

## Working Agreements
### Branching & Reviews
- Follow Git Flow (see ADR-002): `develop` is integration, `main` mirrors production, `feature/<slug>` branches merge via PRs.
- Hotfixes fork from `main` and are back-merged into `develop` after release.
- Every PR references at least one ADR or creates a draft ADR when making non-trivial architectural/process choices.

### Versioning & Tags
- SemVer (`MAJOR.MINOR.PATCH`). Breaking API/storage changes bump `MAJOR`, backward-compatible feature sets bump `MINOR`, and fixes/ops tweaks bump `PATCH`.
- Tags are annotated in the format `vX.Y.Z` and applied on `main` after the release branch is merged.
- Next HA-ready release is expected to become `v0.2.0` once multi-node scheduling and Redis Sentinel integration land.

### Changelog & Tooling
- Maintain `CHANGELOG.md` manually with the Keep a Changelog format. Use `git-chglog --next-tag vX.Y.Z` to hydrate entries, then reconcile wording before committing.
- Keep a `.chglog/config.yml` (to be added) that encodes section ordering so generators stay deterministic.
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
- [ADR-001](docs/adr/001-architecture-and-storage.md): Redis-based MVP and gRPC surface.
- [ADR-002](docs/adr/002-gitflow-and-versioning.md): Git Flow adoption, SemVer policy, and continued Redis usage during MVP.

## Milestones
- **M0 – v0.1.0 (Unreleased):** Redis-backed queue with gRPC API and demo server/worker paths (current focus).
- **M1 – v0.2.0 (Planned, HA focus):**
	- Promote scheduler loop into a dedicated service with leader election so multiple API pods can share the delay queue safely.
	- Introduce Redis Sentinel or Cluster + client-side failover wiring inside `internal/storage/redis`.
	- Implement `JobStore.Remove` via ID->payload hash + Lua script to keep cancellation consistent.
	- Shard queues per topic (e.g., `ddq:tasks:{topic}`) and update configs/docs to describe routing.
	- Extend CI/CD with the release-readiness gate (done) and add regression suites for failover scenarios.

## Change Journal
- **2026-01-08:** Introduced Git Flow + SemVer governance (ADR-002), drafted CHANGELOG scaffolding for `v0.1.0`, and published this development log for AI contributors.
- **2026-01-08:** Wired git-chglog config/template, added release-readiness automation to CI, and enumerated HA milestone tasks for `v0.2.0` planning.
