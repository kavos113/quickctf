package service

import (
	"context"
	"log"

	"github.com/kavos113/quickctf/ctf-server/usecase"
	pb "github.com/kavos113/quickctf/gen/go/api/server/v1"
)

type ClientChallengeService struct {
	pb.UnimplementedClientChallengeServiceServer
	usecase *usecase.ClientChallengeUsecase
}

func NewClientChallengeService(usecase *usecase.ClientChallengeUsecase) *ClientChallengeService {
	return &ClientChallengeService{
		usecase: usecase,
	}
}

func (s *ClientChallengeService) GetChallenges(ctx context.Context, req *pb.GetChallengesRequest) (*pb.GetChallengesResponse, error) {
	challenges, err := s.usecase.GetChallenges(ctx)
	if err != nil {
		log.Printf("Failed to get challenges: %v", err)
		return &pb.GetChallengesResponse{
			ErrorMessage: "failed to get challenges",
		}, nil
	}

	pbChallenges := make([]*pb.Challenge, 0, len(challenges))
	for _, c := range challenges {
		pbChallenges = append(pbChallenges, &pb.Challenge{
			Name:        c.Name,
			Description: c.Description,
			Flag:        c.Flag, // フラグは空文字列になっている
			Points:      int32(c.Points),
			Genre:       c.Genre,
		})
	}

	return &pb.GetChallengesResponse{
		Challenges: pbChallenges,
	}, nil
}

func (s *ClientChallengeService) SubmitFlag(ctx context.Context, req *pb.SubmitFlagRequest) (*pb.SubmitFlagResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get user ID from context: %v", err)
		return &pb.SubmitFlagResponse{
			ErrorMessage: "authentication required",
		}, nil
	}

	isCorrect, pointsAwarded, err := s.usecase.SubmitFlag(
		ctx,
		userID,
		req.Submission.ChallengeId,
		req.Submission.SubmittedFlag,
	)
	if err != nil {
		log.Printf("Failed to submit flag: %v", err)
		return &pb.SubmitFlagResponse{
			ErrorMessage: "failed to submit flag",
		}, nil
	}

	return &pb.SubmitFlagResponse{
		Correct:       isCorrect,
		PointsAwarded: int32(pointsAwarded),
	}, nil
}

func (s *ClientChallengeService) StartInstance(ctx context.Context, req *pb.StartInstanceRequest) (*pb.StartInstanceResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get user ID from context: %v", err)
		return &pb.StartInstanceResponse{
			ErrorMessage: "authentication required",
		}, nil
	}

	err = s.usecase.StartInstance(ctx, userID, req.ChallengeId)
	if err != nil {
		log.Printf("Failed to start instance: %v", err)
		return &pb.StartInstanceResponse{
			ErrorMessage: "failed to start instance",
		}, nil
	}

	return &pb.StartInstanceResponse{}, nil
}

func (s *ClientChallengeService) StopInstance(ctx context.Context, req *pb.StopInstanceRequest) (*pb.StopInstanceResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get user ID from context: %v", err)
		return &pb.StopInstanceResponse{
			ErrorMessage: "authentication required",
		}, nil
	}

	err = s.usecase.StopInstance(ctx, userID, req.ChallengeId)
	if err != nil {
		log.Printf("Failed to stop instance: %v", err)
		return &pb.StopInstanceResponse{
			ErrorMessage: "failed to stop instance",
		}, nil
	}

	return &pb.StopInstanceResponse{}, nil
}

func (s *ClientChallengeService) GetInstanceStatus(ctx context.Context, req *pb.GetInstanceStatusRequest) (*pb.GetInstanceStatusResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get user ID from context: %v", err)
		return &pb.GetInstanceStatusResponse{
			ErrorMessage: "authentication required",
		}, nil
	}

	status, err := s.usecase.GetInstanceStatus(ctx, userID, req.ChallengeId)
	if err != nil {
		log.Printf("Failed to get instance status: %v", err)
		return &pb.GetInstanceStatusResponse{
			ErrorMessage: "failed to get instance status",
		}, nil
	}

	return &pb.GetInstanceStatusResponse{
		Status: string(status),
	}, nil
}
