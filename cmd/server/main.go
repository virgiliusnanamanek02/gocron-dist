package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/virgiliusnanamanek02/gocron-dist/internal/cluster"
	"github.com/virgiliusnanamanek02/gocron-dist/internal/hash"
	"github.com/virgiliusnanamanek02/gocron-dist/internal/scheduler"
	"github.com/virgiliusnanamanek02/gocron-dist/internal/storage"
	"github.com/virgiliusnanamanek02/gocron-dist/pkg/api"
	"google.golang.org/grpc"
)

func main() {

	// 1. Get configuration from flags
	grpcPort := flag.Int("grpc-port", 50051, "Port for gRPC API")
	nodeName := flag.String("name", "", "Unique name for this node")
	port := flag.Int("port", 7946, "Port for Gossip Protocol")
	joinAddr := flag.String("join", "", "Address of another node to join the cluster (optional)")
	flag.Parse()

	if *nodeName == "" {
		log.Fatal("Node name is required! Example: -name=node-1")
	}

	// 2. Initialize Consistent Hashing
	ring := hash.NewConsistent()
	ring.AddNode(*nodeName) // Add self to the ring

	dbDir := fmt.Sprintf("data_%s", *nodeName)
	store, _ := storage.NewStore(dbDir)
	defer store.Close()

	// 3. Initialize Scheduler Engine
	engine := scheduler.NewEngine()
	engine.Storage = store

	// 4. Initialize Cluster (Memberlist)
	// We provide callbacks: if a node Joins/Leaves, update the ring!
	c, err := cluster.NewCluster(
		*nodeName,
		*port,
		*grpcPort,
		*joinAddr,
		func(name string) {
			fmt.Printf("\n[Cluster] Node %s joined.\n", name)
			ring.AddNode(name)
		},
		func(name string) {
			fmt.Printf("\n[Cluster] Node %s left.\n", name)
			// (Optional) Implement ring.RemoveNode if needed
		},
	)
	if err != nil {
		log.Fatalf("Failed to create cluster: %v", err)
	}

	server := &api.Server{
		Engine:   engine,
		Ring:     ring,
		NodeName: *nodeName,
		Cluster:  c,
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	api.RegisterSchedulerServiceServer(grpcServer, server)

	// Start gRPC server
	fmt.Printf("[gRPC] Server running on port %d\n", *grpcPort)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	oldJobs, _ := store.GetAllJobs()
	for _, j := range oldJobs {
		engine.AddJob(j)
		fmt.Printf("[Recovery] Loaded job %s from disk\n", j.ID)
	}

	// 5. Run Engine in Goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go engine.Run(ctx)

	// 6. Simulation: Add job every 10 seconds
	// Here "Magic" happens: Engine only executes if ring.GetNode(jobID) == nodeName
	go func() {
		counter := 0
		for {
			counter++
			jobID := fmt.Sprintf("job-%d", counter)

			// Check ownership
			owner := ring.GetNode(jobID)
			fmt.Printf("[Check] %s owned by %s\n", jobID, owner)

			if owner == *nodeName {
				engine.AddJob(&scheduler.Job{
					ID:      jobID,
					Payload: fmt.Sprintf("Data for %s", jobID),
					NextRun: time.Now().Add(5 * time.Second),
				})
			}
			time.Sleep(10 * time.Second)
		}
	}()

	// Wait for stop signal (Ctrl+C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("Shutting down node...")
}
