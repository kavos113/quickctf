module github.com/kavos113/quickctf

go 1.25.6

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	google.golang.org/grpc/cmd/protoc-gen-go-grpc
	google.golang.org/protobuf/cmd/protoc-gen-go
)

require (
	connectrpc.com/connect v1.19.1 // indirect
	golang.org/x/net v0.43.0 // indirect
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.6.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
