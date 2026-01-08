// Package queue 实现了延迟队列的核心业务逻辑。
// 适用场景：处理 gRPC 请求，参数校验，并调用存储层。
package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	pb "github.com/AkikoAkaki/distributed-delay-queue/api/proto"
	"github.com/AkikoAkaki/distributed-delay-queue/internal/common/errno"
	"github.com/AkikoAkaki/distributed-delay-queue/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service 实现 pb.DelayQueueServiceServer 接口。
type Service struct {
	pb.UnimplementedDelayQueueServiceServer
	store storage.JobStore
}

// NewService 创建一个新的 Service 实例。
func NewService(store storage.JobStore) *Service {
	return &Service{
		store: store,
	}
}

// Enqueue 处理任务提交请求。
func (s *Service) Enqueue(ctx context.Context, req *pb.EnqueueRequest) (*pb.EnqueueResponse, error) {
	// 1. 参数校验 (工业级必须)
	if req.Topic == "" || req.Payload == "" {
		return nil, status.Error(codes.InvalidArgument, errno.ErrInvalidParam.Message)
	}
	if req.DelaySeconds < 0 {
		return nil, status.Error(codes.InvalidArgument, "delay_seconds must be >= 0")
	}

	// 2. 生成 Task ID (如果客户端没传)
	taskID := req.Id
	if taskID == "" {
		taskID = uuid.New().String()
	}

	// 3. 构造核心实体 Task
	task := &pb.Task{
		Id:          taskID,
		Topic:       req.Topic,
		Payload:     req.Payload,
		ExecuteTime: time.Now().Add(time.Duration(req.DelaySeconds) * time.Second).Unix(),
	}

	// 4. 调用存储层
	if err := s.store.Add(ctx, task); err != nil {
		// 这里应该记录 Log，为了简洁省略
		return &pb.EnqueueResponse{
			Success:      false,
			ErrorMessage: "failed to store task",
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.EnqueueResponse{
		Success: true,
		Id:      taskID,
	}, nil
}

// Retrieve 目前 Worker 是直接连 Redis 的，这个接口暂时留空，或者是给外部 Admin 用的。
func (s *Service) Retrieve(ctx context.Context, req *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Retrieve not implemented yet")
}

func (s *Service) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Delete not implemented yet")
}
