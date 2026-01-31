package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/kavos113/quickctf/ctf-runner/service"
	pb "github.com/kavos113/quickctf/gen/go/api/runner/v1"
	"github.com/kavos113/quickctf/lib/logger"
)

func main() {
	port := os.Getenv("RUNNER_PORT")
	if port == "" {
		port = "50052"
	}

	registryURL := os.Getenv("CTF_REGISTRY_URL")
	if registryURL == "" {
		registryURL = "localhost:5000"
	}

	log.Printf("CTF Registry URL: %s", registryURL)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	loggingInterceptor := logger.NewLoggingInterceptor("ctf-runner")

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor.Unary(),
		),
		grpc.ChainStreamInterceptor(
			loggingInterceptor.Stream(),
		),
	)

	runnerService := service.NewRunnerService(registryURL)
	pb.RegisterRunnerServiceServer(grpcServer, runnerService)

	reflection.Register(grpcServer)

	log.Printf("Runner service listening on port %s", port)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		runnerService.Cleanup()
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
