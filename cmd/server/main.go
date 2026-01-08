package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/AkikoAkaki/distributed-delay-queue/api/proto"
	"github.com/AkikoAkaki/distributed-delay-queue/internal/queue"
	"github.com/AkikoAkaki/distributed-delay-queue/internal/storage/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. 初始化配置 (Hardcode for MVP, 实际应读取 config.yaml)
	redisAddr := "localhost:6379"
	port := ":9090"

	// 2. 初始化基础设施
	store := redis.NewStore(redisAddr)

	// 3. 初始化 gRPC Server
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	svc := queue.NewService(store)
	pb.RegisterDelayQueueServiceServer(s, svc)

	// 开启 gRPC Reflection (方便调试工具如 grpcurl 使用)
	reflection.Register(s)

	// 4. 启动服务 (Goroutine)
	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 5. 优雅退出 (Graceful Shutdown)
	// 监听系统信号：Ctrl+C (SIGINT) 或 kill (SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gRPC server...")
	s.GracefulStop()
	log.Println("Server stopped")
}
