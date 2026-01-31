package service

import (
	"context"
	"log"

	"connectrpc.com/connect"

	"github.com/kavos113/quickctf/ctf-server/usecase"
	pb "github.com/kavos113/quickctf/gen/go/api/server/v1"
	"github.com/kavos113/quickctf/gen/go/api/server/v1/serverv1connect"
)

type ClientChallengeService struct {
	serverv1connect.UnimplementedClientChallengeServiceHandler
	usecase *usecase.ClientChallengeUsecase
}

func NewClientChallengeService(usecase *usecase.ClientChallengeUsecase) *ClientChallengeService {
	return &ClientChallengeService{
		usecase: usecase,
	}
}

func (s *ClientChallengeService) GetChallenges(ctx context.Context, req *connect.Request[pb.GetChallengesRequest]) (*connect.Response[pb.GetChallengesResponse], error) {
	challenges, err := s.usecase.GetChallenges(ctx)
	if err != nil {
		log.Printf("Failed to get challenges: %v", err)
		return connect.NewResponse(&pb.GetChallengesResponse{
			ErrorMessage: "failed to get challenges",
		}), nil
	}

	pbChallenges := make([]*pb.Challenge, 0, len(challenges))
	for _, c := range challenges {
		pbChallenges = append(pbChallenges, &pb.Challenge{
			ChallengeId: c.ChallengeID,
			Name:        c.Name,
			Description: c.Description,
			Flag:        c.Flag, // フラグは空文字列になっている
			Points:      int32(c.Points),
			Genre:       c.Genre,
		})
	}

	return connect.NewResponse(&pb.GetChallengesResponse{
		Challenges: pbChallenges,
	}), nil
}

func (s *ClientChallengeService) SubmitFlag(ctx context.Context, req *connect.Request[pb.SubmitFlagRequest]) (*connect.Response[pb.SubmitFlagResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get user ID from context: %v", err)
		return connect.NewResponse(&pb.SubmitFlagResponse{
			ErrorMessage: "authentication required",
		}), nil
	}

	isCorrect, pointsAwarded, err := s.usecase.SubmitFlag(
		ctx,
		userID,
		req.Msg.Submission.ChallengeId,
		req.Msg.Submission.SubmittedFlag,
	)
	if err != nil {
		log.Printf("Failed to submit flag: %v", err)
		return connect.NewResponse(&pb.SubmitFlagResponse{
			ErrorMessage: "failed to submit flag",
		}), nil
	}

	return connect.NewResponse(&pb.SubmitFlagResponse{
		Correct:       isCorrect,
		PointsAwarded: int32(pointsAwarded),
	}), nil
}

func (s *ClientChallengeService) StartInstance(ctx context.Context, req *connect.Request[pb.StartInstanceRequest]) (*connect.Response[pb.StartInstanceResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get user ID from context: %v", err)
		return connect.NewResponse(&pb.StartInstanceResponse{
			ErrorMessage: "authentication required",
		}), nil
	}

	host, port, err := s.usecase.StartInstance(ctx, userID, req.Msg.ChallengeId)
	if err != nil {
		log.Printf("Failed to start instance: %v", err)
		return connect.NewResponse(&pb.StartInstanceResponse{
			ErrorMessage: "failed to start instance",
		}), nil
	}

	return connect.NewResponse(&pb.StartInstanceResponse{
		Host: host,
		Port: port,
	}), nil
}

func (s *ClientChallengeService) StopInstance(ctx context.Context, req *connect.Request[pb.StopInstanceRequest]) (*connect.Response[pb.StopInstanceResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get user ID from context: %v", err)
		return connect.NewResponse(&pb.StopInstanceResponse{
			ErrorMessage: "authentication required",
		}), nil
	}

	err = s.usecase.StopInstance(ctx, userID, req.Msg.ChallengeId)
	if err != nil {
		log.Printf("Failed to stop instance: %v", err)
		return connect.NewResponse(&pb.StopInstanceResponse{
			ErrorMessage: "failed to stop instance",
		}), nil
	}

	return connect.NewResponse(&pb.StopInstanceResponse{}), nil
}

func (s *ClientChallengeService) GetInstanceStatus(ctx context.Context, req *connect.Request[pb.GetInstanceStatusRequest]) (*connect.Response[pb.GetInstanceStatusResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get user ID from context: %v", err)
		return connect.NewResponse(&pb.GetInstanceStatusResponse{
			ErrorMessage: "authentication required",
		}), nil
	}

	status, host, port, err := s.usecase.GetInstanceStatus(ctx, userID, req.Msg.ChallengeId)
	if err != nil {
		log.Printf("Failed to get instance status: %v", err)
		return connect.NewResponse(&pb.GetInstanceStatusResponse{
			ErrorMessage: "failed to get instance status",
		}), nil
	}

	var pbStatus pb.GetInstanceStatusResponse_Status
	switch status {
	case "STATUS_RUNNING":
		pbStatus = pb.GetInstanceStatusResponse_STATUS_RUNNING
	case "STATUS_STOPPED":
		pbStatus = pb.GetInstanceStatusResponse_STATUS_STOPPED
	case "STATUS_DESTROYED":
		pbStatus = pb.GetInstanceStatusResponse_STATUS_DESTROYED
	default:
		pbStatus = pb.GetInstanceStatusResponse_STATUS_UNSPECIFIED
	}

	return connect.NewResponse(&pb.GetInstanceStatusResponse{
		Status: pbStatus,
		Host:   host,
		Port:   port,
	}), nil
}
