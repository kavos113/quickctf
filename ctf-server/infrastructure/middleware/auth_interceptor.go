package middleware

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type contextKey string

const (
	sessionContextKey contextKey = "session"
	userIDContextKey  contextKey = "user_id"
)

type AuthInterceptor struct {
	sessionRepo domain.SessionRepository
	publicMethods map[string]bool
}

func NewAuthInterceptor(sessionRepo domain.SessionRepository) *AuthInterceptor {
	publicMethods := map[string]bool{
		"/api.server.v1.UserAuthService/Register": true,
		"/api.server.v1.UserAuthService/Login":    true,
	}

	return &AuthInterceptor{
		sessionRepo:   sessionRepo,
		publicMethods: publicMethods,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if i.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		token, err := getTokenFromMetadata(ctx)
		if err != nil {
			log.Printf("Failed to get token: %v", err)
			return nil, status.Errorf(codes.Unauthenticated, "authentication required")
		}

		session, err := i.sessionRepo.FindByToken(ctx, token)
		if err != nil {
			if err == domain.ErrSessionNotFound {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token")
			}
			log.Printf("Failed to find session: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to validate token")
		}

		if session.IsExpired() {
			i.sessionRepo.Delete(ctx, token)
			return nil, status.Errorf(codes.Unauthenticated, "session expired")
		}

		ctx = context.WithValue(ctx, sessionContextKey, session)
		ctx = context.WithValue(ctx, userIDContextKey, session.UserID)

		return handler(ctx, req)
	}
}

func getTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata not found")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization token not found")
	}

	token := tokens[0]
	if len(token) > 7 && token[:7] == "Bearer " {
		return token[7:], nil
	}

	return token, nil
}
