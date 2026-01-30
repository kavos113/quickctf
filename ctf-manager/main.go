package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/kavos113/quickctf/ctf-manager/repository"
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

	// データベース接続設定
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "ctf_manager_db"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)

	// データベース接続
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 接続確認
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Printf("Connected to database: %s", dbName)

	// リポジトリ初期化
	repo := repository.NewMySQLInstanceRepository(db)

	// テーブルスキーマ初期化
	if err := repo.InitSchema(ctx); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	log.Printf("Manager service starting on port %s", port)
	log.Printf("Connected runners: %v", runnerURLs)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	managerService, err := service.NewManagerService(runnerURLs, repo)
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
