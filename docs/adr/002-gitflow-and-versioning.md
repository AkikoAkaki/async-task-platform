# 002 - Git Flow, Versioning, and Storage Strategy (Draft)

- **Date:** 2026-01-08
- **Status:** Draft

## Context
- The distributed delay queue is evolving from a single-node demo into a multi-contributor platform that must support delayed job ingestion, scheduling, and future high-availability releases.
- We already rely on Redis Sorted Sets plus Lua scripts for atomic scheduling (see ADR-001), and we expose a gRPC contract so clients and workers share the same typed surface.
- Upcoming features (topic sharding, scheduler services, HA Redis) will require disciplined branching, semantic releases, and reproducible change documentation for both humans and AI contributors.

## Decision
1. **Adopt Git Flow for collaboration.**
   - `main` stays releasable and mirrors production.
   - `develop` collects integration work; feature branches fork from `develop` and merge back via reviewed PRs.
   - Release branches (e.g. `release/v0.1.0`) stabilize milestones, while hotfix branches fork from `main` for urgent patches.
2. **Use Semantic Versioning (SemVer).**
   - `MAJOR` for breaking API/storage changes (e.g. redesigning the gRPC contract or storage schema).
   - `MINOR` for backward-compatible feature additions (e.g. HA scheduler, topic sharding).
   - `PATCH` for bug fixes or operational tweaks.
3. **Continue the Redis-backed layered architecture for the MVP.**
   - Redis ZSet + Lua offers deterministic ordering, millisecond latency, and straightforward local reproducibility.
   - The `storage.JobStore` abstraction insulates handlers/schedulers from backend changes, so we can swap in disk-backed engines later without contract churn.
4. **Document every notable process/storage decision through ADRs and ensure changelog entries map to ADR IDs when relevant.**

## Consequences
- Contributors have a predictable branch map, making reviews and CI automation easier.
- Releases can be tagged with confidence because SemVer expresses the blast radius of each cut.
- Redis remains the default store until a new ADR supersedes this choice; future backends must satisfy the `JobStore` contract and update ADRs accordingly.
- ADRs, the changelog, and the new `DEVELOPMENT_LOG.md` become the canonical trail for AI-driven changes, reducing re-discovery cost.

## Implementation Notes
- **Branch naming:** `feature/<slug>`, `bugfix/<slug>`, `release/vX.Y.Z`, `hotfix/vX.Y.Z`.
- **Tags:** annotated tags in the form `vX.Y.Z` (see Tag Rules section in `DEVELOPMENT_LOG.md`).
- **Changelog:** use `git-chglog` with a repository-specific config (e.g. `.chglog/config.yml`). Generated sections should merge into `CHANGELOG.md` under the appropriate SemVer heading, preserving manual context.
- **ADR workflow:** start each major decision as `Draft`, link to related PRs, and move to `Accepted` once merged into `develop`. Superseded decisions reference the replacing ADR.
