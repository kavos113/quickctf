package service

import (
	"context"
	"log"

	"connectrpc.com/connect"

	"github.com/kavos113/quickctf/ctf-server/domain"
	"github.com/kavos113/quickctf/ctf-server/usecase"
	pb "github.com/kavos113/quickctf/gen/go/api/server/v1"
	"github.com/kavos113/quickctf/gen/go/api/server/v1/serverv1connect"
)

type UserAuthService struct {
	serverv1connect.UnimplementedUserAuthServiceHandler
	usecase *usecase.UserAuthUsecase
}

func NewUserAuthService(usecase *usecase.UserAuthUsecase) *UserAuthService {
	return &UserAuthService{
		usecase: usecase,
	}
}

func (s *UserAuthService) Register(ctx context.Context, req *connect.Request[pb.RegisterRequest]) (*connect.Response[pb.RegisterResponse], error) {
	userID, err := s.usecase.Register(ctx, req.Msg.Username, req.Msg.Password)
	if err != nil {
		log.Printf("Register failed: %v", err)

		errorMsg := "registration failed"
		if err == domain.ErrUserAlreadyExists {
			errorMsg = "username already exists"
		}

		return connect.NewResponse(&pb.RegisterResponse{
			UserId:       "",
			ErrorMessage: errorMsg,
		}), nil
	}

	log.Printf("User registered: %s", userID)
	return connect.NewResponse(&pb.RegisterResponse{
		UserId:       userID,
		ErrorMessage: "",
	}), nil
}

func (s *UserAuthService) Login(ctx context.Context, req *connect.Request[pb.LoginRequest]) (*connect.Response[pb.LoginResponse], error) {
	token, err := s.usecase.Login(ctx, req.Msg.Username, req.Msg.Password)
	if err != nil {
		log.Printf("Login failed: %v", err)

		errorMsg := "login failed"
		if err == domain.ErrInvalidPassword {
			errorMsg = "invalid username or password"
		}

		return connect.NewResponse(&pb.LoginResponse{
			Token:        "",
			ErrorMessage: errorMsg,
		}), nil
	}

	log.Printf("User logged in: %s", req.Msg.Username)
	return connect.NewResponse(&pb.LoginResponse{
		Token:        token,
		ErrorMessage: "",
	}), nil
}

func (s *UserAuthService) Logout(ctx context.Context, req *connect.Request[pb.LogoutRequest]) (*connect.Response[pb.LogoutResponse], error) {
	session, err := getSessionFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get session from context: %v", err)
		return connect.NewResponse(&pb.LogoutResponse{
			ErrorMessage: "logout failed",
		}), nil
	}

	if err := s.usecase.Logout(ctx, session.Token); err != nil {
		log.Printf("Logout failed: %v", err)
		return connect.NewResponse(&pb.LogoutResponse{
			ErrorMessage: "logout failed",
		}), nil
	}

	log.Printf("User logged out: %s", session.UserID)
	return connect.NewResponse(&pb.LogoutResponse{
		ErrorMessage: "",
	}), nil
}
