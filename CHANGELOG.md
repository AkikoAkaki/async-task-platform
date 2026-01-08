# Changelog

All notable changes to this project will be tracked here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and [Semantic Versioning](https://semver.org/). Use `git-chglog` with a repo-specific config (e.g. `.chglog/config.yml`) to regenerate sections, then review and merge the output into this file instead of blindly overwriting manual context.

## [0.1.0] - Unreleased
### Added
- Redis Sorted Set `JobStore` with Lua-based atomic pop to deliver the initial delayed-task scheduling pipeline.
- `DelayQueueService` gRPC contract plus the `cmd/server` and `cmd/worker` reference flows that demonstrate enqueue-and-retrieve behaviour.

### Notes
- Run `git-chglog --next-tag v0.1.0` to preview commits for this milestone before promoting it to a dated release section.

## [0.0.1] - 2026-01-07
### Build
- fix Dockerfile and align structure with coding standards.

### Chore
- initialize project skeleton and dev environment.

### Feat
- implement Redis storage with Lua script.
- define API contract and storage interface.

### Fix
- hard code Redis image version in CI configuration to fix CI problem.
