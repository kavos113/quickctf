package service

import (
	"context"

	"github.com/kavos113/quickctf/ctf-server/domain"
	"github.com/kavos113/quickctf/ctf-server/usecase"
	pb "github.com/kavos113/quickctf/gen/go/api/server/v1"
)

type AdminService struct {
	pb.UnimplementedAdminServiceServer
	adminUsecase *usecase.AdminServiceUsecase
}

func NewAdminService(adminUsecase *usecase.AdminServiceUsecase) *AdminService {
	return &AdminService{
		adminUsecase: adminUsecase,
	}
}

func (s *AdminService) CreateChallenge(ctx context.Context, req *pb.CreateChallengeRequest) (*pb.CreateChallengeResponse, error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return &pb.CreateChallengeResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	challenge := &domain.Challenge{
		Name:        req.Challenge.Name,
		Description: req.Challenge.Description,
		Flag:        req.Challenge.Flag,
		Points:      int(req.Challenge.Points),
		Genre:       req.Challenge.Genre,
	}

	challengeID, err := s.adminUsecase.CreateChallenge(ctx, challenge)
	if err != nil {
		return &pb.CreateChallengeResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.CreateChallengeResponse{
		ChallengeId: challengeID,
	}, nil
}

func (s *AdminService) UpdateChallenge(ctx context.Context, req *pb.UpdateChallengeRequest) (*pb.UpdateChallengeResponse, error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return &pb.UpdateChallengeResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	challenge := &domain.Challenge{
		Name:        req.Challenge.Name,
		Description: req.Challenge.Description,
		Flag:        req.Challenge.Flag,
		Points:      int(req.Challenge.Points),
		Genre:       req.Challenge.Genre,
	}

	err = s.adminUsecase.UpdateChallenge(ctx, req.ChallengeId, challenge)
	if err != nil {
		return &pb.UpdateChallengeResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.UpdateChallengeResponse{}, nil
}

func (s *AdminService) UploadChallengeImage(ctx context.Context, req *pb.UploadChallengeImageRequest) (*pb.UploadChallengeImageResponse, error) {
	return &pb.UploadChallengeImageResponse{
		ErrorMessage: "not implemented yet",
	}, nil
}

func (s *AdminService) DeleteChallenge(ctx context.Context, req *pb.DeleteChallengeRequest) (*pb.DeleteChallengeResponse, error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return &pb.DeleteChallengeResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	err = s.adminUsecase.DeleteChallenge(ctx, req.ChallengeId)
	if err != nil {
		return &pb.DeleteChallengeResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.DeleteChallengeResponse{}, nil
}

func (s *AdminService) ListChallenges(ctx context.Context, req *pb.ListChallengesRequest) (*pb.ListChallengesResponse, error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return &pb.ListChallengesResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	challenges, err := s.adminUsecase.ListChallenges(ctx)
	if err != nil {
		return &pb.ListChallengesResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	pbChallenges := make([]*pb.Challenge, 0, len(challenges))
	for _, c := range challenges {
		pbChallenges = append(pbChallenges, &pb.Challenge{
			Name:        c.Name,
			Description: c.Description,
			Flag:        c.Flag,
			Points:      int32(c.Points),
			Genre:       c.Genre,
		})
	}

	return &pb.ListChallengesResponse{
		Challenges: pbChallenges,
	}, nil
}
