package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/kavos113/quickctf/ctf-manager/service"
	pb "github.com/kavos113/quickctf/gen/go/api/manager/v1"
)

func main() {
	port := os.Getenv("MANAGER_PORT")
	if port == "" {
		port = "50050"
	}

	runnersEnv := os.Getenv("RUNNER_URLS")
	if runnersEnv == "" {
		runnersEnv = "localhost:50052"
	}
	runnerURLs := strings.Split(runnersEnv, ",")

	log.Printf("Manager service starting on port %s", port)
	log.Printf("Connected runners: %v", runnerURLs)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	managerService, err := service.NewManagerService(runnerURLs)
	if err != nil {
		log.Fatalf("failed to create manager service: %v", err)
	}
	pb.RegisterRunnerServiceServer(grpcServer, managerService)

	reflection.Register(grpcServer)

	log.Printf("Manager service listening on port %s", port)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		managerService.Cleanup()
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
