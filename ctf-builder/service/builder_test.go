package service

import (
	"context"
	"testing"

	"io"
	"log"
	"net"

	pb "github.com/kavos113/quickctf/gen/go/api/builder/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterBuilderServiceServer(s, NewBuilderService())
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestBuilderService_BuildImage(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet", 
		grpc.WithContextDialer(bufDialer), 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewBuilderServiceClient(conn)

	// テスト用の簡単なDockerfileを含むtar（ダミーデータ）
	// 実際のテストではtarアーカイブを作成する必要があります
	req := &pb.BuildImageRequest{
		ImageTag:  "test:latest",
		SourceTar: []byte("dummy tar data"),
	}

	stream, err := client.BuildImage(ctx, req)
	if err != nil {
		t.Fatalf("BuildImage failed: %v", err)
	}

	// レスポンスストリームを読み取る
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Logf("Error receiving stream: %v", err)
			break
		}

		switch r := resp.Response.(type) {
		case *pb.BuildImageResponse_LogLine:
			t.Logf("Log: %s", r.LogLine)
		case *pb.BuildImageResponse_Result:
			t.Logf("Result - Status: %s, ImageID: %s, Error: %s", 
				r.Result.Status, r.Result.ImageId, r.Result.ErrorMessage)
			// ダミーデータなので失敗することを期待
			if r.Result.Status != "failed" {
				t.Errorf("Expected failed status for dummy data, got: %s", r.Result.Status)
			}
		}
	}
}
