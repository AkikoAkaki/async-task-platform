# DelayQueueService API

The queue exposes a gRPC service defined in `api/proto/queue.proto`. This document summarizes the RPC surface and shows sample payloads.

## Service Definition
```
service DelayQueueService {
  rpc Enqueue(EnqueueRequest) returns (EnqueueResponse);
  rpc Retrieve(RetrieveRequest) returns (RetrieveResponse);
  rpc Delete(DeleteRequest) returns (DeleteResponse);
}
```

### Messages
- **Task**: `{ id, topic, payload, execute_time }`
- **EnqueueRequest**: `{ topic, payload, delay_seconds, id? }`
- **RetrieveRequest**: `{ topic, batch_size }`
- **DeleteRequest**: `{ id }`

## Typical Calls
### Enqueue
```
grpcurl -plaintext -d '{
  "topic": "order-cancel",
  "payload": "{\"order_id\":1024}",
  "delay_seconds": 300
}' localhost:9090 api.queue.DelayQueueService/Enqueue
```
- The server generates `id` when the field is omitted.
- The Redis store schedules the task by `now + delay_seconds`.

### Retrieve
```
grpcurl -plaintext -d '{
  "topic": "order-cancel",
  "batch_size": 10
}' localhost:9090 api.queue.DelayQueueService/Retrieve
```
- Returns due tasks ordered by `execute_time`.
- Workers should ack each task before fetching the next batch to avoid at-least-once duplicates during restarts.

### Delete
```
grpcurl -plaintext -d '{ "id": "task-001" }' \
  localhost:9090 api.queue.DelayQueueService/Delete
```
- Currently delegated to `JobStore.Remove`. The Redis MVP still returns `not implemented` so clients must tolerate this response.

## Code Generation
Run the `proto` target after editing `queue.proto`:
```
make proto
```
This invokes `protoc` with `--go_out` and `--go-grpc_out`, producing `queue.pb.go` and `queue_grpc.pb.go` under `api/proto`.

## Validation Guidelines
- Keep `topic` as a non-empty ASCII string to simplify key derivation.
- Treat `payload` as opaque JSON text; downstream workers own the schema.
- Guard against negative `delay_seconds`; implementations can clamp or reject values before persistence.
- Clamp `batch_size` to a sane upper bound (for example 100) to protect Redis from large atomic pops.

## Error Handling Guidelines
Implementations should translate domain failures into gRPC status codes:
- Invalid user input -> `codes.InvalidArgument`.
- Redis or storage outages -> `codes.Internal` with a terse message.
- Unsupported paths (such as `Delete` in the current MVP) -> `codes.Unimplemented` so clients can fall back gracefully.

## Versioning Strategy
- Follow semantic versioning for the proto package via git tags.
- Additive fields prefer `reserved` declarations when old fields are removed.
- Backward incompatible changes require revving the `package` or introducing a new service name.
