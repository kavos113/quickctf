## 環境変数

| 変数名 | 説明 | デフォルト値 |
|--------|------|------------|
| `RUNNER_PORT` | gRPCサーバーのポート番号 | `50052` |
| `CTF_REGISTRY_URL` | CTF Registryのアドレス | `localhost:5000` |
| `DOCKER_HOST` | Dockerデーモンのソケット | `/var/run/docker.sock` |
| `MIN_OPEN_PORT` | 開放するポートの最小値 | | 
| `MAX_OPEN_PORT` | 開放するポートの最大値 | |
| `INTERNAL_CONTAINER_PORT` | コンテナ側がexposeするポート | 80 |
