package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/AkikoAkaki/distributed-delay-queue/internal/storage/redis"
)

func main() {
	redisAddr := "localhost:6379"
	store := redis.NewStore(redisAddr)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Worker started, polling for tasks...")

	// 使用 WaitGroup 保证退出时处理完当前 Loop
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second) // 轮询间隔 1秒
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// 收到停止信号，退出循环
				return
			case <-ticker.C:
				// 1. 拉取任务
				tasks, err := store.GetReady(ctx, "default", 10)
				if err != nil {
					log.Printf("Error polling tasks: %v", err)
					continue
				}

				// 2. 执行任务 (MVP: 仅打印)
				if len(tasks) > 0 {
					log.Printf("--- Processed %d tasks ---", len(tasks))
					for _, t := range tasks {
						// 工业级：这里应该扔给一个 Worker Pool 线程池去并发执行，而不是串行阻塞
						log.Printf("[EXECUTE] TaskID: %s, Payload: %s, Delay: %ds", 
							t.Id, t.Payload, time.Now().Unix()-t.ExecuteTime)
					}
				}
			}
		}
	}()

	// 优雅退出监听
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Worker shutting down...")
	cancel()  // 通知 Loop 停止
	wg.Wait() // 等待 Loop 彻底结束
	log.Println("Worker stopped")
}
