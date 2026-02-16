# gocron-dist

**gocron-dist** is a distributed job scheduler written in Go. It is designed to be resilient, scalable, and persistent, solving the limitations of traditional single-node cron jobs.

## Features

- **Distributed Architecture**: Runs on multiple nodes; if one node fails, the cluster continues to function.
- **Persistence**: Jobs are saved to disk using [PebbleDB](https://github.com/cockroachdb/pebble), ensuring no data loss during restarts.
- **Clustering**: Uses [Memberlist](https://github.com/hashicorp/memberlist) (Gossip protocol) for node discovery and failure detection.
- **Horizontal Scaling**: Jobs are distributed across nodes using **Consistent Hashing**.
- **gRPC API**: Submit and manage jobs programmatically via gRPC.

## Architecture

The system consists of several key components:

- **Engine**: The core scheduler that manages the priority queue of jobs and triggers execution.
- **Cluster**: Handles node membership and failure detection.
- **Hash Ring**: Determines which node is responsible for a specific job ID.
- **Storage**: Persistent storage layer for jobs.

## Installation

```bash
go get github.com/virgiliusnanamanek02/gocron-dist
```

## Usage

### Running the Server

```bash
# Start the first node
go run cmd/server/main.go -port 8001 -grpc-port 9001 -node node-1

# Start a second node and join the cluster
go run cmd/server/main.go -port 8002 -grpc-port 9002 -node node-2 -join localhost:8001
```

### Scheduling a Job (Client)

```bash
go run cmd/client/main.go -server localhost:9001 -id "job-1" -schedule "2023-12-01T10:00:00Z" -payload "Send Email"
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
