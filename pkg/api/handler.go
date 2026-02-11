package api

import (
	"context"
	"fmt"

	"github.com/virgiliusnanamanek02/gocron-dist/internal/hash"
	"github.com/virgiliusnanamanek02/gocron-dist/internal/scheduler"
)

type Server struct {
	UnimplementedSchedulerServiceServer
	Engine   *scheduler.Engine
	Ring     *hash.Consistent
	NodeName string
}

func (s *Server) AddJob(ctx context.Context, req *AddJobRequest) (*AddJobResponse, error) {
	// 1. Tentukan siapa pemilik job berdasarkan ID-nya
	owner := s.Ring.GetNode(req.Id)

	// 2. Jika pemiliknya bukan node ini, beri tahu client
	if owner != s.NodeName {
		return &AddJobResponse{
			Success:      false,
			Message:      fmt.Sprintf("Job ini milik node %s", owner),
			AssignedNode: owner,
		}, nil
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
