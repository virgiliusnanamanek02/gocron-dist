package api

import (
	"context"
	"fmt"

	"github.com/virgiliusnanamanek02/gocron-dist/internal/cluster"
	"github.com/virgiliusnanamanek02/gocron-dist/internal/hash"
	"github.com/virgiliusnanamanek02/gocron-dist/internal/scheduler"
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
	// 1. Tentukan siapa pemilik job berdasarkan ID-nya
	owner := s.Ring.GetNode(req.Id)

	// 2. Jika pemiliknya bukan node ini, forward request!
	if owner != s.NodeName {
		// Cari alamat gRPC node pemilik
		addr, err := s.Cluster.GetNodeGrpcAddress(owner)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve owner %s: %v", owner, err)
		}

		fmt.Printf("[Forward] Forwarding job %s to %s (%s)\n", req.Id, owner, addr)

		// Dial ke node pemilik
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to owner %s: %v", owner, err)
		}
		defer conn.Close()

		client := NewSchedulerServiceClient(conn)
		return client.AddJob(ctx, req)
	}

	// 3. Jika milik sendiri, masukkan ke Engine
	s.Engine.AddJob(&scheduler.Job{
		ID:      req.Id,
		Payload: req.Payload,
		NextRun: req.ScheduleTime.AsTime(),
	})

	return &AddJobResponse{
		Success:      true,
		Message:      "Job berhasil dijadwalkan secara lokal",
		AssignedNode: s.NodeName,
	}, nil
}
