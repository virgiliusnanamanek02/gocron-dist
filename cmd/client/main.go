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
	// Flag untuk menentukan mau nembak ke node mana
	target := flag.String("target", "localhost:50051", "Alamat gRPC server")
	jobID := flag.String("id", "job-1", "ID unik untuk job")
	payload := flag.String("msg", "Hello dari Client!", "Isi pesan job")
	delay := flag.Int("delay", 5, "Delay eksekusi dalam detik")
	flag.Parse()

	// 1. Konek ke server gRPC
	conn, err := grpc.Dial(*target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Gagal konek: %v", err)
	}
	defer conn.Close()

	client := api.NewSchedulerServiceClient(conn)

	// 2. Siapkan request dengan waktu eksekusi di masa depan
	scheduleTime := time.Now().Add(time.Duration(*delay) * time.Second)
	req := &api.AddJobRequest{
		Id:           *jobID,
		Payload:      *payload,
		ScheduleTime: timestamppb.New(scheduleTime),
	}

	// 3. Kirim request
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.AddJob(ctx, req)
	if err != nil {
		log.Fatalf("Gagal kirim job: %v", err)
	}

	// 4. Cek hasil
	if res.Success {
		log.Printf("SUKSES: %s (Diterima oleh: %s)", res.Message, res.AssignedNode)
	} else {
		log.Printf("GAGAL: %s (Harusnya ke: %s)", res.Message, res.AssignedNode)
	}
}
