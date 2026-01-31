package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/kavos113/quickctf/ctf-server/infrastructure/client"
	"github.com/kavos113/quickctf/ctf-server/infrastructure/repository"
	"github.com/kavos113/quickctf/ctf-server/interface/middleware"
	"github.com/kavos113/quickctf/ctf-server/interface/service"
	"github.com/kavos113/quickctf/ctf-server/usecase"
	"github.com/kavos113/quickctf/gen/go/api/server/v1/serverv1connect"
	"github.com/kavos113/quickctf/lib/logger"
)

func main() {
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
	loggingInterceptor := logger.NewConnectLoggingInterceptor("ctf-server")

	interceptors := connect.WithInterceptors(authInterceptor, loggingInterceptor)

	log.Printf("CTF server starting on port %s", port)

	mux := http.NewServeMux()

	path, handler := serverv1connect.NewUserAuthServiceHandler(userAuthService, interceptors)
	mux.Handle(path, handler)

	path, handler = serverv1connect.NewAdminAuthServiceHandler(adminAuthService, interceptors)
	mux.Handle(path, handler)

	path, handler = serverv1connect.NewAdminServiceHandler(adminService, interceptors)
	mux.Handle(path, handler)

	path, handler = serverv1connect.NewClientChallengeServiceHandler(clientChallengeService, interceptors)
	mux.Handle(path, handler)

	corsHandler := corsMiddleware(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: h2c.NewHandler(corsHandler, &http2.Server{}),
	}

	log.Printf("CTF server listening on port %s", port)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to serve: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, Connect-Protocol-Version")
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
