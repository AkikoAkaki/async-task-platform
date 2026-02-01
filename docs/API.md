# TaskService API Reference

The Async Task Platform exposes a gRPC service for task management. This document provides the complete API reference with examples.

## Service Definition

```protobuf
service DelayQueueService {
  // Submit a delayed task for future execution
  rpc Enqueue(EnqueueRequest) returns (EnqueueResponse);
  
  // Retrieve due tasks (typically called by workers)
  rpc Retrieve(RetrieveRequest) returns (RetrieveResponse);
  
  // Cancel a pending task by ID
  rpc Delete(DeleteRequest) returns (DeleteResponse);
}
```

> **Note**: The service is currently named `DelayQueueService` for backward compatibility. It will be renamed to `TaskService` in a future major version.

## Messages

### Task

The core task entity:

```protobuf
message Task {
  string id = 1;           // Unique identifier
  string topic = 2;        // Logical grouping (e.g., "order-cancel", "email-send")
  string payload = 3;      // Business data as JSON string
  int64  execute_time = 4; // Scheduled execution time (Unix timestamp)
  int32  retry_count = 5;  // Current retry attempt (0 = first attempt)
  int32  max_retries = 6;  // Maximum retries before moving to DLQ
  int64  created_at = 7;   // Task creation timestamp
}
```

### EnqueueRequest / EnqueueResponse

```protobuf
message EnqueueRequest {
  string topic = 1;           // Required: business topic
  string payload = 2;         // Required: JSON payload
  int64  delay_seconds = 3;   // Required: delay before execution (>= 0)
  string id = 4;              // Optional: client-provided ID for idempotency
  int32  max_retries = 5;     // Optional: custom retry limit (default: 3)
}

message EnqueueResponse {
  bool   success = 1;         // Whether the task was enqueued
  string id = 2;              // Assigned task ID
  string error_message = 3;   // Error details if success=false
}
```

### RetrieveRequest / RetrieveResponse

```protobuf
message RetrieveRequest {
  string topic = 1;           // Topic to retrieve from
  int32  batch_size = 2;      // Maximum tasks to return (capped at 100)
}

message RetrieveResponse {
  repeated Task tasks = 1;    // List of due tasks
}
```

### DeleteRequest / DeleteResponse

```protobuf
message DeleteRequest {
  string id = 1;              // Task ID to cancel
}

message DeleteResponse {
  bool success = 1;           // Whether deletion succeeded
}
```

## API Examples

### Prerequisites

Install `grpcurl` for command-line testing:

```powershell
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### Enqueue: Submit a Delayed Task

**Basic usage:**

```powershell
grpcurl -plaintext -d '{
  "topic": "order-cancel",
  "payload": "{\"order_id\": 1024, \"user_id\": 42}",
  "delay_seconds": 1800
}' localhost:9090 api.queue.DelayQueueService/Enqueue
```

**Response:**

```json
{
  "success": true,
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

**With custom ID (for idempotency):**

```powershell
grpcurl -plaintext -d '{
  "topic": "order-cancel",
  "payload": "{\"order_id\": 1024}",
  "delay_seconds": 1800,
  "id": "order-1024-cancel"
}' localhost:9090 api.queue.DelayQueueService/Enqueue
```

**With custom retry limit:**

```powershell
grpcurl -plaintext -d '{
  "topic": "critical-job",
  "payload": "{}",
  "delay_seconds": 60,
  "max_retries": 5
}' localhost:9090 api.queue.DelayQueueService/Enqueue
```

### Retrieve: Fetch Due Tasks

> **Status**: Currently returns `UNIMPLEMENTED`. Workers use internal `FetchAndHold` method.

```powershell
grpcurl -plaintext -d '{
  "topic": "order-cancel",
  "batch_size": 10
}' localhost:9090 api.queue.DelayQueueService/Retrieve
```

### Delete: Cancel a Pending Task

> **Status**: Currently returns `UNIMPLEMENTED`. Will be implemented in Phase 1.

```powershell
grpcurl -plaintext -d '{
  "id": "order-1024-cancel"
}' localhost:9090 api.queue.DelayQueueService/Delete
```

## Error Handling

The API uses standard gRPC status codes:

| Code | Meaning | Example |
|------|---------|---------|
| `OK` | Success | Task enqueued |
| `INVALID_ARGUMENT` | Bad input | Empty topic, negative delay |
| `NOT_FOUND` | Resource missing | Delete non-existent task |
| `INTERNAL` | Server error | Redis connection failed |
| `UNIMPLEMENTED` | Feature not ready | Delete API in MVP |

**Example error response:**

```json
{
  "code": 3,
  "message": "topic and payload are required",
  "details": []
}
```

## Validation Rules

| Field | Rule |
|-------|------|
| `topic` | Required, non-empty ASCII string |
| `payload` | Required, valid JSON string |
| `delay_seconds` | Required, must be >= 0 |
| `batch_size` | Capped at 100 to prevent large atomic pops |
| `id` | If provided, must be unique; duplicates create new task |

## Code Generation

After modifying `api/proto/queue.proto`, regenerate Go code:

```powershell
make proto
```

This requires:
- `protoc` (Protocol Buffer compiler)
- `protoc-gen-go` (Go code generator)
- `protoc-gen-go-grpc` (gRPC code generator)

## SDK Usage (Go)

```go
import (
    pb "github.com/AkikoAkaki/async-task-platform/api/proto"
    "google.golang.org/grpc"
)

func main() {
    conn, _ := grpc.Dial("localhost:9090", grpc.WithInsecure())
    client := pb.NewDelayQueueServiceClient(conn)
    
    resp, err := client.Enqueue(context.Background(), &pb.EnqueueRequest{
        Topic:        "order-cancel",
        Payload:      `{"order_id": 1024}`,
        DelaySeconds: 1800,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Task ID: %s\n", resp.Id)
}
```

## Versioning

- Follow semantic versioning for proto package via git tags
- Additive fields are backward compatible
- Breaking changes require bumping major version or new service name
- Use `reserved` declarations when removing fields

