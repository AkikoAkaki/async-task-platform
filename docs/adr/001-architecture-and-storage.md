# 1. Architecture and Storage Selection

Date: 2026-01-04  
Status: Accepted  
Updated: 2026-02-01

## Context

We are building a distributed asynchronous task scheduling platform that needs to support:

1. **Delayed Execution**: Tasks execute at a precise future time
2. **Periodic Scheduling**: Tasks repeat on a schedule (future phase)
3. **Workflow Orchestration**: Multi-step task pipelines (future phase)

The system must be highly available, low latency, and capable of handling millions of scheduled tasks.

## Decision

### Language: Go

**Rationale:**
- Excellent concurrency primitives (goroutines, channels)
- Low memory footprint suitable for infrastructure services
- Strong ecosystem for gRPC and Redis clients
- First-class support for distributed systems patterns

### Protocol: gRPC with Protobuf

**Rationale:**
- Strongly-typed contracts with code generation
- Better performance than JSON/HTTP for internal services
- Built-in streaming support for future worker implementations
- Industry standard for microservice communication

### Storage: Redis (MVP Phase)

**Primary Data Structure: Sorted Set (ZSet)**

```
Key:    ddq:tasks
Score:  execute_time (Unix timestamp)
Member: JSON-serialized Task
```

**Supporting Structures:**
- `ddq:running` (Hash): Tasks currently being processed
- `ddq:dlq` (List): Dead Letter Queue for failed tasks

**Atomicity:** Lua scripts guarantee atomic operations:
- `FetchAndHold`: Atomically move tasks from pending to running
- `Ack/Nack`: Atomically complete or retry tasks
- `Recover`: Atomically detect and recover timeout tasks

## Consequences

### Advantages
- **Performance**: Redis provides sub-millisecond latency for scheduling operations
- **Simplicity**: ZSet naturally supports time-based ordering with O(log N) insertion
- **Atomicity**: Lua scripts prevent race conditions in distributed environments
- **Operational Maturity**: Redis is well-understood with proven HA solutions (Sentinel, Cluster)

### Limitations
- **Memory Bound**: Cannot store billions of pending tasks (typically < 10M practical)
- **Persistence Risk**: Default RDB snapshot may lose tasks on crash (mitigate with AOF)
- **No Native Priority**: Priority must be encoded in score (e.g., high bits for priority)

### Future Evolution Path

| Phase | Storage Enhancement |
|-------|---------------------|
| MVP | Single Redis instance with Lua scripts |
| HA | Redis Sentinel for automatic failover |
| Scale | Redis Cluster with topic-based sharding |
| Volume | Hybrid: Redis for hot tasks + PostgreSQL/RocksDB for cold storage |

## References

- [Redis Sorted Sets](https://redis.io/docs/data-types/sorted-sets/)
- [Distributed Locks with Redis](https://redis.io/docs/manual/patterns/distributed-locks/)
- [Exactly-Once Delivery Patterns](https://exactly-once.github.io/)
