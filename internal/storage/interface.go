// Package storage 定义了持久化存储层的抽象接口。
// 适用场景：解耦业务逻辑与具体的 Redis/MySQL 实现。
package storage

import (
	"context"

	pb "github.com/AkikoAkaki/distributed-delay-queue/api/proto"
)

// JobStore 定义了任务存储的所有行为。
// 实现该接口的结构体应当处理具体的 Redis 命令或 SQL 语句。
type JobStore interface {
	// Add 将任务添加到存储引擎中。
	// 参数 ctx: 上下文，用于超时控制。
	// 参数 task: 待存储的任务模型。
	Add(ctx context.Context, task *pb.Task) error

	// GetReady 获取并移除已到期的任务。
	// 参数 topic: 任务主题。
	// 参数 limit: 单词拉取的最大数量。
	// 返回值: 任务列表。
	GetReady(ctx context.Context, topic string, limit int64) ([]*pb.Task, error)

	// Remove 删除指定任务。
	Remove(ctx context.Context, id string) error
}
