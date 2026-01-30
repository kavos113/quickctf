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
	// メタデータからトークンを取得する必要があるが、簡易実装のため
	// passwordフィールドを"token:code"形式で受け取る
	// TODO: gRPC metadataからトークンを取得する実装に変更
	
	// 仮実装：passwordを"token:activationCode"として解析
	// 本来はmetadataからtokenを取得し、passwordにactivationCodeを入れる
	log.Printf("Admin activation requested")
	
	// 実際の実装では、metadataからトークンを取得し、
	// req.Passwordをactivation codeとして使う
	// ここではプレースホルダーとして空の応答を返す
	
	return &pb.AdminLoginResponse{}, nil
}

// AdminLogout はadmin権限を解除
func (s *AdminAuthService) AdminLogout(ctx context.Context, req *pb.AdminLogoutRequest) (*pb.AdminLogoutResponse, error) {
	// メタデータからトークンを取得してadmin権限を解除
	// TODO: 実装が必要
	
	log.Printf("Admin deactivation requested")
	return &pb.AdminLogoutResponse{}, nil
}
