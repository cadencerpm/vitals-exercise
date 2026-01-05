package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cadence-vitals-interview/internal/api"
	"cadence-vitals-interview/internal/app"
	vitalsv1 "cadence-vitals-interview/proto/vitals/v1"
	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", ":50051", "gRPC listen address")
	flag.Parse()

	store := app.NewInMemoryStore()
	pubsub := app.NewPubSub()
	service := app.NewService(store, pubsub)
	worker := app.NewAlertWorker(pubsub, store, 16)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go worker.Run(ctx)

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", *addr, err)
	}

	grpcServer := grpc.NewServer()
	vitalsv1.RegisterVitalsServiceServer(grpcServer, api.NewServer(service))

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
		pubsub.Close()
		store.Close()
	}()

	log.Printf("vitals server listening on %s", *addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("grpc server stopped: %v", err)
	}
}
