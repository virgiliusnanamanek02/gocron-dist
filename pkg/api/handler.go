package api

import (
	"context"
	"fmt"
	"log"

	"github.com/vnmchuo/gocron-dist/internal/cluster"
	"github.com/vnmchuo/gocron-dist/internal/hash"
	"github.com/vnmchuo/gocron-dist/internal/scheduler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	UnimplementedSchedulerServiceServer
	Engine   *scheduler.Engine
	Ring     *hash.Consistent
	Cluster  *cluster.Cluster
	NodeName string
}

func (s *Server) AddJob(ctx context.Context, req *AddJobRequest) (*AddJobResponse, error) {
	// 1. Determine who owns the job based on its ID
	owner := s.Ring.GetNode(req.Id)

	// 2. If the owner is not this node, forward the request!
	if owner != s.NodeName {
		// Find the gRPC address of the owner node
		addr, err := s.Cluster.GetNodeGrpcAddress(owner)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve owner %s: %v", owner, err)
		}

		log.Printf("[Forward] Forwarding job %s to %s (%s)\n", req.Id, owner, addr)

		// Dial the owner node
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to owner %s: %v", owner, err)
		}
		defer conn.Close()

		client := NewSchedulerServiceClient(conn)
		return client.AddJob(ctx, req)
	}

	// 3. If it belongs to this node, add it to the Engine
	s.Engine.AddJob(&scheduler.Job{
		ID:      req.Id,
		Payload: req.Payload,
		NextRun: req.ScheduleTime.AsTime(),
	})

	return &AddJobResponse{
		Success:      true,
		Message:      "Job scheduled successfully locally",
		AssignedNode: s.NodeName,
	}, nil
}
