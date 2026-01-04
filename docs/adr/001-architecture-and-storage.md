# 1. Architecture and Storage Selection

Date: 2026-01-04
Status: Accepted

## Context
我们需要构建一个分布式的延迟队列系统，要求高可用、低延迟，并支持海量任务的定时调度。

## Decision
1. **Language:** 使用 Go。原因：并发处理能力强 (Goroutines)，内存占用低，适合中间件开发。
2. **Protocol:** 使用 gRPC (Protobuf)。原因：强类型契约，性能优于 JSON/HTTP，支持代码生成。
3. **Storage:** 早期阶段 (MVP) 使用 **Redis ZSet (Sorted Set)**。
   - **Score:** 任务执行的时间戳。
   - **Member:** 任务 ID。
   - 配合 Redis Lua 脚本保证原子性。

## Consequences
- **Pros:** Redis 性能极高，开发简单，ZSet 天然支持按时间排序的范围查询。
- **Cons:** Redis 内存受限，无法存储海量（如十亿级）积压任务。
- **Future:** 后续考虑引入 "Time Wheel" + 磁盘存储 (LevelDB/MySQL) 来解决存储瓶颈。
