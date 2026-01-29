module github.com/kavos113/quickctf

go 1.25.6

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	github.com/golang/protobuf/protoc-gen-go
	google.golang.org/grpc/cmd/protoc-gen-go-grpc
)

require (
	connectrpc.com/connect v1.19.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.6.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
