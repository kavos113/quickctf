## 環境変数

| 変数名 | 説明 | デフォルト値 |
|--------|------|------------|
| `SERVER_PORT` | gRPCサーバーのポート番号 | `50060` |
| `DB_HOST` | MySQLホスト | `localhost` |
| `DB_PORT` | MySQLポート | `3306` |
| `DB_USER` | MySQLユーザー名 | `root` |
| `DB_PASSWORD` | MySQLパスワード | `password` |
| `DB_NAME` | データベース名 | `ctf_server_db` |
| `SCHEMA_PATH` | スキーマファイルのパス | `../migration/ctf_server_schema.sql` |
| `REDIS_ADDRESS` | Redisのアドレス | `localhost:6379` |
| `REDIS_PASSWORD` | Redisのパスワード | (なし) |
| `MANAGER_ADDRESS` | ctf-managerのアドレス | `localhost:50052` |
| `ADMIN_ACTIVATION_CODE` | 管理者アクティベーションコード | `admin_secret` |

