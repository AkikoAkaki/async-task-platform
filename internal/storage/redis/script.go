package redis

// luaPeekAndRem 实现了分布式延时队列的“消费并删除”原子操作。
// @Logic
// 1. ZRANGEBYSCORE: 基于当前系统时间戳，在有序集合(ZSet)中检索所有已到期的任务 ID。
// 2. ZREM: 同步从 ZSet 中剔除上述命中的任务，防止任务被并发节点重复拉取。
// 3. Return: 将命中的任务 ID 列表返回给调用方进行后续的业务处理。
//
// @Constraints
// - 原子性保障：通过 Lua 脚本执行，确保读取与删除之间不被其他命令插入。
// - 性能限制：调用方需合理控制 ARGV[2] (limit)，避免大批量删除导致 Redis 阻塞。
//
// @Parameters
// KEYS[1] - string: 延时队列的 ZSet 键名 (e.g., "dq_tasks_zset")
// ARGV[1] - int64 : 当前 Unix 时间戳 (Score)，用于判定任务是否到期
// ARGV[2] - int   : 单词拉取的最大任务数量 (Limit)，用于流量削峰
//
// @Returns
// table: 返回包含任务 Payload (ID) 的数组；若无到期任务则返回空 Table。
const luaPeekAndRem = `
local key = KEYS[1]
local max_score = ARGV[1]
local limit = ARGV[2]

-- 1. 检索所有 Score 小于等于当前时间戳的任务
local tasks = redis.call('ZRANGEBYSCORE', key, 0, max_score, 'LIMIT', 0, limit)

if #tasks > 0 then
    -- 2. 执行原子删除，利用 unpack 将 table 展开为多参数模式
    redis.call('ZREM', key, unpack(tasks))
    return tasks
else
    -- 3. 显式返回空 table，对应 Go 中的空切片
    return {}
end
`
