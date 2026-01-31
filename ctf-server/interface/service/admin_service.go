package service

import (
	"context"

	"connectrpc.com/connect"

	"github.com/kavos113/quickctf/ctf-server/domain"
	"github.com/kavos113/quickctf/ctf-server/usecase"
	pb "github.com/kavos113/quickctf/gen/go/api/server/v1"
	"github.com/kavos113/quickctf/gen/go/api/server/v1/serverv1connect"
)

type AdminService struct {
	serverv1connect.UnimplementedAdminServiceHandler
	adminUsecase *usecase.AdminServiceUsecase
}

func NewAdminService(adminUsecase *usecase.AdminServiceUsecase) *AdminService {
	return &AdminService{
		adminUsecase: adminUsecase,
	}
}

func (s *AdminService) CreateChallenge(ctx context.Context, req *connect.Request[pb.CreateChallengeRequest]) (*connect.Response[pb.CreateChallengeResponse], error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return connect.NewResponse(&pb.CreateChallengeResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	challenge := &domain.Challenge{
		Name:        req.Msg.Challenge.Name,
		Description: req.Msg.Challenge.Description,
		Flag:        req.Msg.Challenge.Flag,
		Points:      int(req.Msg.Challenge.Points),
		Genre:       req.Msg.Challenge.Genre,
	}

	challengeID, err := s.adminUsecase.CreateChallenge(ctx, challenge)
	if err != nil {
		return connect.NewResponse(&pb.CreateChallengeResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&pb.CreateChallengeResponse{
		ChallengeId: challengeID,
	}), nil
}

func (s *AdminService) UpdateChallenge(ctx context.Context, req *connect.Request[pb.UpdateChallengeRequest]) (*connect.Response[pb.UpdateChallengeResponse], error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return connect.NewResponse(&pb.UpdateChallengeResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	challenge := &domain.Challenge{
		Name:        req.Msg.Challenge.Name,
		Description: req.Msg.Challenge.Description,
		Flag:        req.Msg.Challenge.Flag,
		Points:      int(req.Msg.Challenge.Points),
		Genre:       req.Msg.Challenge.Genre,
	}

	err = s.adminUsecase.UpdateChallenge(ctx, req.Msg.Challenge.ChallengeId, challenge)
	if err != nil {
		return connect.NewResponse(&pb.UpdateChallengeResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&pb.UpdateChallengeResponse{}), nil
}

func (s *AdminService) UploadChallengeImage(ctx context.Context, req *connect.Request[pb.UploadChallengeImageRequest]) (*connect.Response[pb.UploadChallengeImageResponse], error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return connect.NewResponse(&pb.UploadChallengeImageResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	err = s.adminUsecase.UploadChallengeImage(ctx, req.Msg.ChallengeId, req.Msg.ImageData)
	if err != nil {
		return connect.NewResponse(&pb.UploadChallengeImageResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&pb.UploadChallengeImageResponse{}), nil
}

func (s *AdminService) DeleteChallenge(ctx context.Context, req *connect.Request[pb.DeleteChallengeRequest]) (*connect.Response[pb.DeleteChallengeResponse], error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return connect.NewResponse(&pb.DeleteChallengeResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	err = s.adminUsecase.DeleteChallenge(ctx, req.Msg.ChallengeId)
	if err != nil {
		return connect.NewResponse(&pb.DeleteChallengeResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&pb.DeleteChallengeResponse{}), nil
}

func (s *AdminService) ListChallenges(ctx context.Context, req *connect.Request[pb.ListChallengesRequest]) (*connect.Response[pb.ListChallengesResponse], error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return connect.NewResponse(&pb.ListChallengesResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	challenges, err := s.adminUsecase.ListChallenges(ctx)
	if err != nil {
		return connect.NewResponse(&pb.ListChallengesResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	pbChallenges := make([]*pb.Challenge, 0, len(challenges))
	for _, c := range challenges {
		pbChallenges = append(pbChallenges, &pb.Challenge{
			ChallengeId: c.ChallengeID,
			Name:        c.Name,
			Description: c.Description,
			Flag:        c.Flag,
			Points:      int32(c.Points),
			Genre:       c.Genre,
		})
	}

	return connect.NewResponse(&pb.ListChallengesResponse{
		Challenges: pbChallenges,
	}), nil
}
