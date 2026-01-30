package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/kavos113/quickctf/gen/go/api/builder/v1"
)

type BuilderClient struct {
	client pb.BuilderServiceClient
	conn   *grpc.ClientConn
}

func NewBuilderClient() (*BuilderClient, error) {
	builderAddr := os.Getenv("BUILDER_ADDRESS")
	if builderAddr == "" {
		builderAddr = "localhost:50051"
	}

	conn, err := grpc.NewClient(builderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to builder service: %w", err)
	}

	client := pb.NewBuilderServiceClient(conn)

	return &BuilderClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *BuilderClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *BuilderClient) BuildImage(ctx context.Context, imageTag string, sourceTar []byte) (string, error) {
	stream, err := c.client.BuildImage(ctx, &pb.BuildImageRequest{
		ImageTag:  imageTag,
		SourceTar: sourceTar,
	})
	if err != nil {
		return "", fmt.Errorf("failed to call BuildImage: %w", err)
	}

	var imageID string
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to receive response: %w", err)
		}

		switch r := resp.Response.(type) {
		case *pb.BuildImageResponse_LogLine:
			log.Printf("[Builder] %s", r.LogLine)
		case *pb.BuildImageResponse_Result:
			if r.Result.Status != "success" {
				return "", fmt.Errorf("build failed: %s", r.Result.ErrorMessage)
			}
			imageID = r.Result.ImageId
		}
	}

	if imageID == "" {
		return "", fmt.Errorf("build completed but no image ID returned")
	}

	return imageID, nil
}
