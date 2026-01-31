package service

import (
	"context"
	"log"

	"connectrpc.com/connect"

	"github.com/kavos113/quickctf/ctf-server/usecase"
	pb "github.com/kavos113/quickctf/gen/go/api/server/v1"
	"github.com/kavos113/quickctf/gen/go/api/server/v1/serverv1connect"
)

type AdminAuthService struct {
	serverv1connect.UnimplementedAdminAuthServiceHandler
	usecase *usecase.AdminAuthUsecase
}

func NewAdminAuthService(usecase *usecase.AdminAuthUsecase) *AdminAuthService {
	return &AdminAuthService{
		usecase: usecase,
	}
}

func (s *AdminAuthService) AdminLogin(ctx context.Context, req *connect.Request[pb.AdminLoginRequest]) (*connect.Response[pb.AdminLoginResponse], error) {
	session, err := getSessionFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get session from context: %v", err)
		return connect.NewResponse(&pb.AdminLoginResponse{}), nil
	}

	err = s.usecase.ActivateAdminWithSession(ctx, session, req.Msg.Password)
	if err != nil {
		log.Printf("Failed to activate admin: %v", err)
		return connect.NewResponse(&pb.AdminLoginResponse{}), nil
	}

	log.Printf("Admin activated for user: %s", session.UserID)
	return connect.NewResponse(&pb.AdminLoginResponse{}), nil
}

func (s *AdminAuthService) AdminLogout(ctx context.Context, req *connect.Request[pb.AdminLogoutRequest]) (*connect.Response[pb.AdminLogoutResponse], error) {
	session, err := getSessionFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get session from context: %v", err)
		return connect.NewResponse(&pb.AdminLogoutResponse{}), nil
	}

	err = s.usecase.DeactivateAdminWithSession(ctx, session)
	if err != nil {
		log.Printf("Failed to deactivate admin: %v", err)
		return connect.NewResponse(&pb.AdminLogoutResponse{}), nil
	}

	log.Printf("Admin deactivated for user: %s", session.UserID)
	return connect.NewResponse(&pb.AdminLogoutResponse{}), nil
}
