module github.com/kavos113/quickctf/ctf-server

go 1.25.6

require (
	connectrpc.com/connect v1.18.1
	github.com/go-sql-driver/mysql v1.9.3
	github.com/google/uuid v1.6.0
	github.com/kavos113/quickctf/gen v0.0.0
	github.com/kavos113/quickctf/lib v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.17.3
	golang.org/x/crypto v0.44.0
	golang.org/x/net v0.47.0
	google.golang.org/grpc v1.78.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/kavos113/quickctf/gen => ../gen
	github.com/kavos113/quickctf/lib => ../lib
)
