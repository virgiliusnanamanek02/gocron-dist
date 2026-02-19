# GEMINI.md - gocron-dist Context

This file provides instructional context for Gemini CLI when working on the `gocron-dist` project.

## Project Overview
`gocron-dist` is a distributed, resilient, and persistent job scheduler written in Go. It aims to solve the limitations of single-node cron systems by distributing tasks across a cluster of nodes and ensuring they survive restarts.

### Core Technologies
- **Go 1.24**: Main programming language.
- **PebbleDB**: High-performance key-value store used for job persistence.
- **Memberlist (Hashicorp)**: Implements the Gossip protocol for node discovery and cluster membership.
- **gRPC**: Protocol for the external API and inter-node job forwarding.
- **Consistent Hashing**: CRC32-based ring used to determine job ownership across nodes.

## Architecture & Components

### 1. Scheduler Engine (`internal/scheduler`)
- Uses a **min-heap** priority queue to manage job execution times.
- Supports **recurring jobs** via `RepeatInterval` and `MaxRuns`.
- Supports **best-effort rebalancing**: if a node leaves, surviving nodes re-evaluate their job ownership and forward jobs to new owners.
- Instrumented with **OpenTelemetry** tracing.

### 2. Clustering & Distribution (`internal/cluster` & `internal/hash`)
- **Memberlist**: Nodes discover each other and exchange metadata (like gRPC ports).
- **Consistent Hashing**: Maps Job IDs to specific nodes. Supports dynamic node removal and re-assignment.
- **Forwarding**: Automatically forwards `AddJob` requests to the correct owner node.

### 3. Persistence (`internal/storage`)
- Each node maintains its own **PebbleDB** instance.
- Jobs are persisted before execution and deleted after completion (unless recurring).
- Instrumented with **OpenTelemetry** tracing.

### 4. API (`pkg/api`)
- gRPC service supporting recurring job configuration.
- Instrumented with **OpenTelemetry** tracing.

### 5. Telemetry (`internal/telemetry`)
- OpenTelemetry initialization with stdout exporter (pretty-print).
- Critical path instrumentation: gRPC -> Hash Ring -> Storage -> Queue -> Execution.

## Building and Running

### Prerequisites
- Go 1.24+
- `protoc` (if modifying `.proto` files)

### Key Commands
- **Run a Node**:
  ```bash
  go run cmd/server/main.go -name node-1 -port 7946 -grpc-port 9001
  ```
- **Join a Cluster**:
  ```bash
  go run cmd/server/main.go -name node-2 -port 7947 -grpc-port 9002 -join localhost:7946
  ```
- **Add a Job (Client)**:
  ```bash
  go run cmd/client/main.go -target localhost:9001 -id "my-task" -delay 10 -msg "Hello World"
  ```
- **Test**:
  ```bash
  go test ./...
  ```

## Development Conventions

### Code Structure
- **`cmd/`**: Entry points for server and client.
- **`internal/`**: Core logic that should not be imported by external projects.
- **`pkg/`**: Public API and generated code.
- **`examples/`**: Usage examples for library-style consumption.

### Best Practices
- **Idiomatic Go**: Follow standard Go formatting (`gofmt`) and naming conventions.
- **Concurrency**: Use channels and `sync.Mutex` carefully, especially in the `Engine` and `Consistent` hash ring.
- **Error Handling**: Log errors clearly, especially during gRPC forwarding and storage operations.
- **Persistence**: Always ensure jobs are flushed to PebbleDB (`pebble.Sync`) to guarantee resilience.

## Future Considerations (TODOs)
- **Job Rescheduling**: Currently, jobs are deleted after execution. Support for recurring (cron) jobs is partially defined in the `Job` struct but not fully implemented in the engine.
- **Dynamic Rebalancing**: Implementing automatic job transfer when a node joins or leaves the cluster.
- **Telemetry**: Enhance `internal/telemetry` for better observability.
