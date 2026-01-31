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

func (s *AdminService) GetChallenge(ctx context.Context, req *connect.Request[pb.GetChallengeRequest]) (*connect.Response[pb.GetChallengeResponse], error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return connect.NewResponse(&pb.GetChallengeResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	challenge, err := s.adminUsecase.GetChallenge(ctx, req.Msg.ChallengeId)
	if err != nil {
		return connect.NewResponse(&pb.GetChallengeResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&pb.GetChallengeResponse{
		Challenge: &pb.Challenge{
			ChallengeId: challenge.ChallengeID,
			Name:        challenge.Name,
			Description: challenge.Description,
			Flag:        challenge.Flag,
			Points:      int32(challenge.Points),
			Genre:       challenge.Genre,
		},
	}), nil
}

func (s *AdminService) ListBuildLogs(ctx context.Context, req *connect.Request[pb.ListBuildLogsRequest]) (*connect.Response[pb.ListBuildLogsResponse], error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return connect.NewResponse(&pb.ListBuildLogsResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	logs, err := s.adminUsecase.ListBuildLogs(ctx, req.Msg.ChallengeId)
	if err != nil {
		return connect.NewResponse(&pb.ListBuildLogsResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	pbLogs := make([]*pb.BuildLogSummary, 0, len(logs))
	for _, l := range logs {
		pbLogs = append(pbLogs, &pb.BuildLogSummary{
			JobId:       l.JobID,
			ChallengeId: l.ChallengeID,
			Status:      l.Status,
			CreatedAt:   l.CreatedAt.Format("2006-01-02 15:04:05"),
			CompletedAt: l.CompletedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return connect.NewResponse(&pb.ListBuildLogsResponse{
		Logs: pbLogs,
	}), nil
}

func (s *AdminService) GetBuildLog(ctx context.Context, req *connect.Request[pb.GetBuildLogRequest]) (*connect.Response[pb.GetBuildLogResponse], error) {
	_, err := requireAdminSession(ctx)
	if err != nil {
		return connect.NewResponse(&pb.GetBuildLogResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	logContent, status, err := s.adminUsecase.GetBuildLog(ctx, req.Msg.JobId)
	if err != nil {
		return connect.NewResponse(&pb.GetBuildLogResponse{
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&pb.GetBuildLogResponse{
		JobId:      req.Msg.JobId,
		LogContent: logContent,
		Status:     status,
	}), nil
}
