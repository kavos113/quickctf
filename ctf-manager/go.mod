module github.com/kavos113/quickctf/ctf-manager

go 1.25.6

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/kavos113/quickctf/gen v0.0.0
	google.golang.org/grpc v1.78.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/kavos113/quickctf/gen => ../gen
