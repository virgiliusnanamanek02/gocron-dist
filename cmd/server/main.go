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

	// 1. Ambil konfigurasi dari flag
	grpcPort := flag.Int("grpc-port", 50051, "Port untuk gRPC API")
	nodeName := flag.String("name", "", "Nama unik untuk node ini")
	port := flag.Int("port", 7946, "Port untuk Gossip Protocol")
	joinAddr := flag.String("join", "", "Alamat node lain untuk bergabung ke cluster (opsional)")
	flag.Parse()

	if *nodeName == "" {
		log.Fatal("Nama node harus diisi! Contoh: -name=node-1")
	}

	// 2. Inisialisasi Consistent Hashing
	ring := hash.NewConsistent()
	ring.AddNode(*nodeName) // Masukkan diri sendiri ke ring

	dbDir := fmt.Sprintf("data_%s", *nodeName)
	store, _ := storage.NewStore(dbDir)
	defer store.Close()

	// 3. Inisialisasi Scheduler Engine
	engine := scheduler.NewEngine()
	engine.Storage = store

	// 4. Inisialisasi Cluster (Memberlist)
	// Kita berikan callback: kalau ada node Join/Leave, update ring-nya!
	c, err := cluster.NewCluster(
		*nodeName,
		*port,
		*grpcPort,
		*joinAddr,
		func(name string) {
			fmt.Printf("\n[Cluster] Node %s bergabung.\n", name)
			ring.AddNode(name)
		},
		func(name string) {
			fmt.Printf("\n[Cluster] Node %s keluar.\n", name)
			// (Opsional) Implementasikan ring.RemoveNode jika diperlukan
		},
	)
	if err != nil {
		log.Fatalf("Gagal membuat cluster: %v", err)
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
	fmt.Printf("[gRPC] Server jalan di port %d\n", *grpcPort)
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

	// 5. Jalankan Engine di Goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go engine.Run(ctx)

	// 6. Simulasi: Tambah job setiap 10 detik
	// Di sini "Magic" terjadi: Engine hanya akan eksekusi jika ring.GetNode(jobID) == nodeName
	go func() {
		counter := 0
		for {
			counter++
			jobID := fmt.Sprintf("job-%d", counter)

			// Cek kepemilikan
			owner := ring.GetNode(jobID)
			fmt.Printf("[Check] %s dimiliki oleh %s\n", jobID, owner)

			if owner == *nodeName {
				engine.AddJob(&scheduler.Job{
					ID:      jobID,
					Payload: fmt.Sprintf("Data untuk %s", jobID),
					NextRun: time.Now().Add(5 * time.Second),
				})
			}
			time.Sleep(10 * time.Second)
		}
	}()

	// Tunggu sinyal stop (Ctrl+C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("Shutting down node...")
}
