package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kavos113/quickctf/ctf-builder/queue"
	"github.com/kavos113/quickctf/ctf-builder/service"
	"github.com/kavos113/quickctf/ctf-builder/storage"
)

func main() {
	registryURL := os.Getenv("CTF_REGISTRY_URL")
	if registryURL == "" {
		registryURL = "localhost:5000"
	}

	log.Printf("CTF Registry URL: %s", registryURL)

	redisClient, err := queue.NewRedisClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	log.Println("Connected to Redis")

	s3Client, err := storage.NewS3Client()
	if err != nil {
		log.Printf("Warning: Failed to connect to S3: %v", err)
		log.Println("Build logs will not be persisted to S3")
		s3Client = nil
	} else {
		log.Println("Connected to S3")
	}

	worker := service.NewBuildWorker(redisClient, s3Client)

	ctx, cancel := context.WithCancel(context.Background())

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
	}()

	worker.Start(ctx)
}
