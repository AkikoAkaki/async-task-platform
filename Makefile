.PHONY: run-server run-worker up down proto lint test

# 启动基础设施 (Redis)
up:
	docker-compose -f deployments/docker-compose.yaml up -d

# 关闭基础设施
down:
	docker-compose -f deployments/docker-compose.yaml down

# 运行 Server (暂时只是个空壳)
run-server:
	go run cmd/server/main.go

# 运行 Worker (暂时只是个空壳)
run-worker:
	go run cmd/worker/main.go

# 生成 Proto 代码 (Phase 4 会用到，先占位)
proto:
	@echo "Generating protobuf code..."
	# protoc command will go here

# 代码检查 (Phase 3 会用到)
lint:
	golangci-lint run

# 运行测试
test:
	go test -v -race ./...
