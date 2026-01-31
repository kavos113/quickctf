package service

import (
	"context"
	"errors"

	"github.com/kavos113/quickctf/ctf-server/domain"
	"github.com/kavos113/quickctf/ctf-server/interface/middleware"
)

func getSessionFromContext(ctx context.Context) (*domain.Session, error) {
	session, ok := ctx.Value(middleware.SessionContextKey).(*domain.Session)
	if !ok {
		return nil, errors.New("session not found in context")
	}
	return session, nil
}

func getUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(middleware.UserIDContextKey).(string)
	if !ok {
		return "", errors.New("user_id not found in context")
	}
	return userID, nil
}

func requireAdminSession(ctx context.Context) (*domain.Session, error) {
	session, err := getSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if !session.IsAdmin {
		return nil, errors.New("admin permission required")
	}

	return session, nil
}
