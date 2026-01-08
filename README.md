<div align="center">

# Distributed Delay Queue

Lightweight, Redis-backed delayed job service with a gRPC surface and room to evolve into a fully distributed scheduler.

</div>

## Highlights
- **Strong contract:** All client and worker interactions go through `DelayQueueService` (Protobuf + gRPC).
- **Atomic scheduling:** Redis Sorted Sets plus Lua scripts ensure that due jobs are popped exactly once per batch.
- **Composable layers:** Storage adapters live behind `internal/storage.JobStore`, so new backends can drop in without disturbing handlers.
- **Infrastructure friendly:** Docker Compose spins up Redis locally; Make targets wrap the common workflows.

## Architecture Snapshot
- Enqueue requests persist tasks with `execute_time` scores inside `ddq:tasks`.
- A scheduler loop (currently embedded in `cmd/server`) polls Redis for due tasks and hands them to workers.
- Workers issue `Retrieve` in batches and process payloads for their topic.
- Cancel operations are wired through the `Delete` RPC; the Redis MVP still returns `not implemented`.
- See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the full component map and scaling notes.

## Repository Layout
| Path | Purpose |
|------|---------|
| `api/proto` | Protobuf definitions and generated stubs. |
| `cmd/server` | Example server process that exercises the Redis store. |
| `cmd/worker` | Worker stub (to be expanded with real execution logic). |
| `config` | Sample YAML configuration (`config.example.yaml`). |
| `deploy` | Docker Compose stack for Redis. |
| `internal/storage` | Storage interfaces plus the Redis implementation. |
| `docs` | Developer setup, API contract, architecture notes, and ADRs. |

## Quick Start
1. **Install prerequisites**
	- Go 1.21+
	- Docker Desktop with WSL2 backend
	- `make` (available via Chocolatey or Scoop)
2. **Clone and configure**
	```powershell
	git clone https://github.com/AkikoAkaki/distributed-delay-queue.git
	cd distributed-delay-queue
	Copy-Item config/config.example.yaml config/config.yaml
	```
3. **Boot Redis**
	```powershell
	make up
	```
4. **Run the demo server**
	```powershell
	make run-server
	```
	The current implementation enqueues a sample task and polls Redis until the job becomes due, showcasing the atomic pop path.
5. **(Optional) Run the worker stub**
	```powershell
	make run-worker
	```
	This target is a placeholder you can extend with custom business logic.
6. **Tear down infrastructure**
	```powershell
	make down
	```

## Make Targets
| Command | Description |
|---------|-------------|
| `make up` / `make down` | Start or stop the Redis dependency via Docker Compose. |
| `make run-server` | Execute `cmd/server` (demo of enqueue + retrieval). |
| `make run-worker` | Execute `cmd/worker` (currently prints a placeholder message). |
| `make proto` | Regenerate Go stubs from `api/proto/queue.proto`. Requires `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`. |
| `make fmt` | Apply `goimports` formatting across the repo. |
| `make lint` | Run `golangci-lint` (depends on `make fmt`). |
| `make test` | Execute the Go test suite with `-race`. |
| `make build-server` / `make build-worker` | Produce binaries in `bin/`. |

## Development Notes
- Keep `config/config.yaml` out of version control; the example file documents every field.
- Generated files (`queue.pb.go`, `queue_grpc.pb.go`) live next to the proto for easy review.
- The Redis store uses JSON for payload serialization right now; switch to Protobuf when memory pressure becomes a concern.
- Use `docs/DEV_SETUP.md` if you need a detailed Windows onboarding checklist.

## Documentation Map
- [docs/API.md](docs/API.md) — RPC shapes, sample `grpcurl` commands, and code-gen hints.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — component responsibilities, runtime flows, scaling FAQs.
- [docs/DEV_SETUP.md](docs/DEV_SETUP.md) — Windows-first environment setup.
- [docs/adr/001-architecture-and-storage.md](docs/adr/001-architecture-and-storage.md) — decision record for the Redis-based MVP.

## Roadmap Ideas
- Flesh out `cmd/server` into a long-running gRPC service with proper config loading.
- Implement a production-grade worker that streams tasks via `Retrieve` instead of the current stub.
- Finish `JobStore.Remove` by introducing an id-to-payload hash and background cleanup.
- Add metrics (Prometheus) and tracing (OpenTelemetry) once the scheduler loop becomes a daemonized component.
