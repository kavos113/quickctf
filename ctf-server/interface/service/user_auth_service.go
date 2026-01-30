package service

import (
	"context"
	"log"

	"github.com/kavos113/quickctf/ctf-server/domain"
	"github.com/kavos113/quickctf/ctf-server/usecase"
	pb "github.com/kavos113/quickctf/gen/go/api/server/v1"
)

type UserAuthService struct {
	pb.UnimplementedUserAuthServiceServer
	usecase *usecase.UserAuthUsecase
}

func NewUserAuthService(usecase *usecase.UserAuthUsecase) *UserAuthService {
	return &UserAuthService{
		usecase: usecase,
	}
}

func (s *UserAuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userID, err := s.usecase.Register(ctx, req.Username, req.Password)
	if err != nil {
		log.Printf("Register failed: %v", err)
		
		errorMsg := "registration failed"
		if err == domain.ErrUserAlreadyExists {
			errorMsg = "username already exists"
		}
		
		return &pb.RegisterResponse{
			UserId:       "",
			ErrorMessage: errorMsg,
		}, nil
	}

	log.Printf("User registered: %s", userID)
	return &pb.RegisterResponse{
		UserId:       userID,
		ErrorMessage: "",
	}, nil
}

func (s *UserAuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := s.usecase.Login(ctx, req.Username, req.Password)
	if err != nil {
		log.Printf("Login failed: %v", err)
		
		errorMsg := "login failed"
		if err == domain.ErrInvalidPassword {
			errorMsg = "invalid username or password"
		}
		
		return &pb.LoginResponse{
			Token:        "",
			ErrorMessage: errorMsg,
		}, nil
	}

	log.Printf("User logged in: %s", req.Username)
	return &pb.LoginResponse{
		Token:        token,
		ErrorMessage: "",
	}, nil
}

func (s *UserAuthService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// コンテキストからセッションを取得
	session, err := getSessionFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get session from context: %v", err)
		return &pb.LogoutResponse{
			ErrorMessage: "logout failed",
		}, nil
	}

	// トークンでログアウト
	if err := s.usecase.Logout(ctx, session.Token); err != nil {
		log.Printf("Logout failed: %v", err)
		return &pb.LogoutResponse{
			ErrorMessage: "logout failed",
		}, nil
	}

	log.Printf("User logged out: %s", session.UserID)
	return &pb.LogoutResponse{
		ErrorMessage: "",
	}, nil
}
