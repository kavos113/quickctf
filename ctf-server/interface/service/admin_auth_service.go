package service

import (
	"context"
	"log"

	"github.com/kavos113/quickctf/ctf-server/usecase"
	pb "github.com/kavos113/quickctf/gen/go/api/server/v1"
)

// AdminAuthService は管理者認証サービスの実装
type AdminAuthService struct {
	pb.UnimplementedAdminAuthServiceServer
	usecase *usecase.AdminAuthUsecase
}

// NewAdminAuthService は新しいAdminAuthServiceを作成
func NewAdminAuthService(usecase *usecase.AdminAuthUsecase) *AdminAuthService {
	return &AdminAuthService{
		usecase: usecase,
	}
}

// AdminLogin はアクティベーションコードでadmin権限を付与
func (s *AdminAuthService) AdminLogin(ctx context.Context, req *pb.AdminLoginRequest) (*pb.AdminLoginResponse, error) {
	// コンテキストからセッションを取得
	session, err := getSessionFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get session from context: %v", err)
		return &pb.AdminLoginResponse{}, nil
	}

	// アクティベーションコードで管理者権限を付与
	err = s.usecase.ActivateAdminWithSession(ctx, session, req.Password)
	if err != nil {
		log.Printf("Failed to activate admin: %v", err)
		return &pb.AdminLoginResponse{}, nil
	}

	log.Printf("Admin activated for user: %s", session.UserID)
	return &pb.AdminLoginResponse{}, nil
}

// AdminLogout はadmin権限を解除
func (s *AdminAuthService) AdminLogout(ctx context.Context, req *pb.AdminLogoutRequest) (*pb.AdminLogoutResponse, error) {
	// コンテキストからセッションを取得
	session, err := getSessionFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get session from context: %v", err)
		return &pb.AdminLogoutResponse{}, nil
	}

	// admin権限を解除
	err = s.usecase.DeactivateAdminWithSession(ctx, session)
	if err != nil {
		log.Printf("Failed to deactivate admin: %v", err)
		return &pb.AdminLogoutResponse{}, nil
	}

	log.Printf("Admin deactivated for user: %s", session.UserID)
	return &pb.AdminLogoutResponse{}, nil
}
