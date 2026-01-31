package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/kavos113/quickctf/ctf-server/infrastructure/client"
	"github.com/kavos113/quickctf/ctf-server/infrastructure/repository"
	"github.com/kavos113/quickctf/ctf-server/interface/middleware"
	"github.com/kavos113/quickctf/ctf-server/interface/service"
	"github.com/kavos113/quickctf/ctf-server/usecase"
	pb "github.com/kavos113/quickctf/gen/go/api/server/v1"
)

func main() {
	ctx := context.Background()

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "50060"
	}

	dbConfig := repository.NewConfigFromEnv()
	db, err := repository.Connect(dbConfig)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := repository.InitSchema(ctx, db, dbConfig.SchemaPath); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	userRepo := repository.NewMySQLUserRepository(db)
	sessionRepo := repository.NewMySQLSessionRepository(db)
	challengeRepo := repository.NewMySQLChallengeRepository(db)
	submissionRepo := repository.NewMySQLSubmissionRepository(db)
	instanceRepo := repository.NewMySQLInstanceRepository(db)

	builderClient, err := client.NewBuilderClient()
	if err != nil {
		log.Fatalf("failed to create builder client: %v", err)
	}
	defer builderClient.Close()

	managerClient, err := client.NewManagerClient()
	if err != nil {
		log.Fatalf("failed to create manager client: %v", err)
	}
	defer managerClient.Close()

	userAuthUsecase := usecase.NewUserAuthUsecase(userRepo, sessionRepo)
	adminAuthUsecase := usecase.NewAdminAuthUsecase(sessionRepo)
	adminServiceUsecase := usecase.NewAdminServiceUsecase(challengeRepo, sessionRepo, builderClient)
	clientChallengeUsecase := usecase.NewClientChallengeUsecase(challengeRepo, submissionRepo, instanceRepo, managerClient)

	userAuthService := service.NewUserAuthService(userAuthUsecase)
	adminAuthService := service.NewAdminAuthService(adminAuthUsecase)
	adminService := service.NewAdminService(adminServiceUsecase)
	clientChallengeService := service.NewClientChallengeService(clientChallengeUsecase)

	authInterceptor := middleware.NewAuthInterceptor(sessionRepo)
	loggingInterceptor := middleware.NewLoggingInterceptor()

	log.Printf("CTF server starting on port %s", port)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor.Unary(),
			authInterceptor.Unary(),
		),
		grpc.ChainStreamInterceptor(
			loggingInterceptor.Stream(),
		),
	)

	pb.RegisterUserAuthServiceServer(grpcServer, userAuthService)
	pb.RegisterAdminAuthServiceServer(grpcServer, adminAuthService)
	pb.RegisterAdminServiceServer(grpcServer, adminService)
	pb.RegisterClientChallengeServiceServer(grpcServer, clientChallengeService)

	reflection.Register(grpcServer)

	log.Printf("CTF server listening on port %s", port)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
