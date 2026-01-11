package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/AkikoAkaki/distributed-delay-queue/api/proto"
	"github.com/AkikoAkaki/distributed-delay-queue/internal/conf" // 引入配置包
	"github.com/AkikoAkaki/distributed-delay-queue/internal/queue"
	"github.com/AkikoAkaki/distributed-delay-queue/internal/storage/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. 加载配置
	// 注意：在 Docker 中 config 可能会 mount 到 /app/config
	// 在本地开发时，可能在 ./config
	cfg, err := conf.Load("./config")
	if err != nil {
		// 如果本地找不到，尝试上一级目录（兼容 go run 在不同目录执行的情况）
		cfg, err = conf.Load("../../config")
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
	}

	log.Printf("Starting %s [%s]...", cfg.App.Name, cfg.App.Env)

	// 2. 初始化 Redis (使用配置)
	store := redis.NewStore(cfg.Redis.Addr)

	// 3. 监听端口 (使用配置)
	addr := fmt.Sprintf(":%d", cfg.Server.GrpcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	svc := queue.NewService(store)
	pb.RegisterDelayQueueServiceServer(s, svc)
	reflection.Register(s)

	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 4. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gRPC server...")
	s.GracefulStop()
	log.Println("Server stopped")
}
