# minicr

minicr is a toy container registry, which supports [OCI Distribution Specs v1.1](https://github.com/opencontainers/distribution-spec).

## 環境変数

### 共通
- `STORAGE_BACKEND`: Blob ストレージバックエンド (`filesystem` または `s3`)。デフォルト: `filesystem`
- `STORE_BACKEND`: メタデータストアバックエンド (`boltdb` または `dynamodb`)。デフォルト: `boltdb`

### Filesystem (デフォルト)
- `STORAGE_PATH`: データ保存先のパス

### S3
- `S3_BUCKET`: S3 バケット名。デフォルト: `ctf-registry`
- `S3_ENDPOINT`: カスタムエンドポイント (MinIO, LocalStack 等)
- `AWS_REGION`: AWS リージョン。デフォルト: `us-east-1`
- `AWS_ACCESS_KEY_ID`: AWS アクセスキー
- `AWS_SECRET_ACCESS_KEY`: AWS シークレットキー

### DynamoDB
- `DYNAMODB_ENDPOINT`: カスタムエンドポイント (LocalStack, DynamoDB Local 等)
- `DYNAMODB_TABLE_PREFIX`: テーブル名のプレフィックス
- `AWS_REGION`: AWS リージョン。デフォルト: `us-east-1`
- `AWS_ACCESS_KEY_ID`: AWS アクセスキー
- `AWS_SECRET_ACCESS_KEY`: AWS シークレットキー