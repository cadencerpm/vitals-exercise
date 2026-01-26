package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cadence-vitals-interview/internal/api"
	"cadence-vitals-interview/internal/app"
	vitalsv1 "cadence-vitals-interview/proto/vitals/v1"
	"google.golang.org/grpc"
)

func main() {
	grpcAddr := flag.String("grpc-addr", ":50051", "gRPC listen address")
	httpAddr := flag.String("http-addr", ":8080", "HTTP listen address for dashboard")
	flag.Parse()

	store := app.NewInMemoryStore()
	pubsub := app.NewPubSub()
	service := app.NewService(store, pubsub)

	// Message queue for patient notifications (5-20 second simulated delay)
	messageQueue := app.NewMessageQueue(5*time.Second, 20*time.Second)
	messageWorker := app.NewMessageWorker(messageQueue)

	// Alert worker with message queue
	worker := app.NewAlertWorker(pubsub, store, 16, messageQueue)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go worker.Run(ctx)
	go messageWorker.Run(ctx)

	// Start gRPC server
	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", *grpcAddr, err)
	}

	grpcServer := grpc.NewServer()
	vitalsv1.RegisterVitalsServiceServer(grpcServer, api.NewServer(service))

	// Start HTTP server for dashboard
	httpServer := api.NewHTTPServer(service, messageQueue)
	httpSrv := &http.Server{
		Addr:    *httpAddr,
		Handler: httpServer.Handler(),
	}

	go func() {
		log.Printf("HTTP dashboard listening on %s", *httpAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpSrv.Shutdown(shutdownCtx)

		pubsub.Close()
		store.Close()
	}()

	log.Printf("gRPC server listening on %s", *grpcAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("grpc server stopped: %v", err)
	}
}
