package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/virgiliusnanamanek02/gocron-dist/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	// Flag to determine which node to hit
	target := flag.String("target", "localhost:50051", "Address of gRPC server")
	jobID := flag.String("id", "job-1", "Unique ID for the job")
	payload := flag.String("msg", "Hello from Client!", "Job message content")
	delay := flag.Int("delay", 5, "Execution delay in seconds")
	flag.Parse()

	// 1. Connect to gRPC server
	conn, err := grpc.NewClient(*target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := api.NewSchedulerServiceClient(conn)

	// 2. Prepare request with future execution time
	scheduleTime := time.Now().Add(time.Duration(*delay) * time.Second)
	req := &api.AddJobRequest{
		Id:           *jobID,
		Payload:      *payload,
		ScheduleTime: timestamppb.New(scheduleTime),
	}

	// 3. Send request
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.AddJob(ctx, req)
	if err != nil {
		log.Fatalf("Failed to send job: %v", err)
	}

	// 4. Check result
	if res.Success {
		log.Printf("SUCCESS: %s (Accepted by: %s)", res.Message, res.AssignedNode)
	} else {
		log.Printf("FAILED: %s (Should be to: %s)", res.Message, res.AssignedNode)
	}
}
