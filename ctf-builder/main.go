package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kavos113/quickctf/ctf-builder/queue"
	"github.com/kavos113/quickctf/ctf-builder/service"
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

	worker := service.NewBuildWorker(redisClient)

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
