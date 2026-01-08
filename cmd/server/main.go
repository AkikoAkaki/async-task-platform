package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/AkikoAkaki/distributed-delay-queue/api/proto"
	"github.com/AkikoAkaki/distributed-delay-queue/internal/storage/redis"
)

func main() {
	// 1. 初始化 Store
	store := redis.NewStore("localhost:6379")
	ctx := context.Background()

	// 2. 模拟生产任务 (延迟 5 秒)
	fmt.Println("--- Producer: Adding Task ---")
	task := &pb.Task{
		Id:          "task-001",
		Payload:     `{"order_id": 1024}`,
		ExecuteTime: time.Now().Add(5 * time.Second).Unix(),
	}
	if err := store.Add(ctx, task); err != nil {
		log.Fatalf("Add failed: %v", err)
	}
	fmt.Println("Task added, waiting for 5 seconds...")

	// 3. 模拟 Worker 轮询
	for i := 0; i < 10; i++ {
		tasks, err := store.GetReady(ctx, "default", 10)
		if err != nil {
			log.Printf("Poll error: %v", err)
		}

		if len(tasks) > 0 {
			fmt.Printf("--- Worker: Got %d tasks! ---\n", len(tasks))
			for _, t := range tasks {
				fmt.Printf("Executing Task: ID=%s, Payload=%s\n", t.Id, t.Payload)
			}
			return // 成功取到，退出
		}

		fmt.Println("Worker: No ready tasks, sleeping 1s...")
		time.Sleep(1 * time.Second)
	}
}