package middleware

import (
	"context"
	"log"

	"connectrpc.com/connect"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type contextKey string

const (
	sessionContextKey contextKey = "session"
	userIDContextKey  contextKey = "user_id"
)

type AuthInterceptor struct {
	sessionRepo   domain.SessionRepository
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

func (i *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := req.Spec().Procedure

		if i.publicMethods[procedure] {
			return next(ctx, req)
		}

		token := req.Header().Get("Authorization")
		if token == "" {
			log.Printf("Authorization header not found")
			return nil, connect.NewError(connect.CodeUnauthenticated, domain.ErrSessionNotFound)
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		session, err := i.sessionRepo.FindByToken(ctx, token)
		if err != nil {
			if err == domain.ErrSessionNotFound {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			log.Printf("Failed to find session: %v", err)
			return nil, connect.NewError(connect.CodeInternal, err)
		}

		if session.IsExpired() {
			i.sessionRepo.Delete(ctx, token)
			return nil, connect.NewError(connect.CodeUnauthenticated, domain.ErrSessionExpired)
		}

		ctx = context.WithValue(ctx, sessionContextKey, session)
		ctx = context.WithValue(ctx, userIDContextKey, session.UserID)

		return next(ctx, req)
	}
}

func (i *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
