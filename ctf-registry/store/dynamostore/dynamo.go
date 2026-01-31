package dynamostore

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/kavos113/quickctf/ctf-registry/manifest"
	"github.com/kavos113/quickctf/ctf-registry/storage"
	"github.com/opencontainers/go-digest"
)

const (
	tableNameTags       = "registry_tags"
	tableNameReferences = "registry_references"
	tableNameBlobs      = "registry_blobs"
)

type Storage struct {
	client      *dynamodb.Client
	tablePrefix string
}

func NewStore() *Storage {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	tablePrefix := os.Getenv("DYNAMODB_TABLE_PREFIX")
	if tablePrefix == "" {
		tablePrefix = ""
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	ctx := context.Background()

	var cfg aws.Config
	var err error

	if endpoint != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err != nil {
			log.Fatalf("failed to load AWS config: %v", err)
		}
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			log.Fatalf("failed to load AWS config: %v", err)
		}
	}

	var client *dynamodb.Client
	if endpoint != "" {
		client = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	} else {
		client = dynamodb.NewFromConfig(cfg)
	}

	s := &Storage{
		client:      client,
		tablePrefix: tablePrefix,
	}

	s.ensureTables(ctx)

	return s
}

func (s *Storage) tableName(base string) string {
	if s.tablePrefix != "" {
		return s.tablePrefix + "_" + base
	}
	return base
}

func (s *Storage) ensureTables(ctx context.Context) {
	tables := []struct {
		name string
		pk   string
		sk   string
	}{
		{tableNameTags, "pk", "sk"},
		{tableNameReferences, "pk", "sk"},
		{tableNameBlobs, "pk", "sk"},
	}

	for _, table := range tables {
		tableName := s.tableName(table.name)
		_, err := s.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err == nil {
			continue
		}

		_, err = s.client.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName: aws.String(tableName),
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String(table.pk),
					KeyType:       types.KeyTypeHash,
				},
				{
					AttributeName: aws.String(table.sk),
					KeyType:       types.KeyTypeRange,
				},
			},
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String(table.pk),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: aws.String(table.sk),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			BillingMode: types.BillingModePayPerRequest,
		})
		if err != nil {
			log.Printf("Warning: failed to create table %s: %v", tableName, err)
		} else {
			log.Printf("Created table %s", tableName)
			time.Sleep(2 * time.Second)
		}
	}
}

type tagItem struct {
	PK     string `dynamodbav:"pk"`
	SK     string `dynamodbav:"sk"`
	Digest string `dynamodbav:"digest"`
}

func (s *Storage) SaveTag(repoName string, d digest.Digest, tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	item := tagItem{
		PK:     fmt.Sprintf("REPO#%s", repoName),
		SK:     fmt.Sprintf("TAG#%s", tag),
		Digest: d.String(),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal tag: %w", storage.ErrStorageFail)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName(tableNameTags)),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to save tag: %w", storage.ErrStorageFail)
	}

	return nil
}

func (s *Storage) ReadTag(repoName string, tag string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName(tableNameTags)),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("REPO#%s", repoName)},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("TAG#%s", tag)},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to read tag: %w", storage.ErrStorageFail)
	}

	if output.Item == nil {
		return "", storage.ErrNotFound
	}

	var item tagItem
	if err := attributevalue.UnmarshalMap(output.Item, &item); err != nil {
		return "", fmt.Errorf("failed to unmarshal tag: %w", storage.ErrStorageFail)
	}

	return item.Digest, nil
}

func (s *Storage) DeleteTag(repoName string, tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.ReadTag(repoName, tag)
	if err != nil {
		return err
	}

	_, err = s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName(tableNameTags)),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("REPO#%s", repoName)},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("TAG#%s", tag)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", storage.ErrStorageFail)
	}

	return nil
}

func (s *Storage) GetTagList(repoName string, limit int, last string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName(tableNameTags)),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: fmt.Sprintf("REPO#%s", repoName)},
			":prefix": &types.AttributeValueMemberS{Value: "TAG#"},
		},
	}

	output, err := s.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", storage.ErrStorageFail)
	}

	tags := make([]string, 0, len(output.Items))
	for _, item := range output.Items {
		var ti tagItem
		if err := attributevalue.UnmarshalMap(item, &ti); err != nil {
			continue
		}
		// Remove "TAG#" prefix
		tag := ti.SK[4:]
		tags = append(tags, tag)
	}

	sort.Strings(tags)

	if last != "" {
		idx := -1
		for i, t := range tags {
			if t == last {
				idx = i
				break
			}
		}
		if idx >= 0 && idx < len(tags)-1 {
			tags = tags[idx+1:]
		}
	}

	if limit > 0 && len(tags) > limit {
		tags = tags[:limit]
	}

	return tags, nil
}

// Reference operations

type referenceItem struct {
	PK          string `dynamodbav:"pk"`
	SK          string `dynamodbav:"sk"`
	Descriptors string `dynamodbav:"descriptors"` // JSON encoded
}

func (s *Storage) AddReference(repoName string, d digest.Digest, desc manifest.Descriptor) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pk := fmt.Sprintf("REPO#%s", repoName)
	sk := fmt.Sprintf("REF#%s", d.String())

	existing, err := s.getReferencesInternal(ctx, pk, sk)
	if err != nil && err != storage.ErrNotFound {
		return err
	}

	existing = append(existing, desc)

	data, err := json.Marshal(existing)
	if err != nil {
		return fmt.Errorf("failed to marshal descriptors: %w", storage.ErrStorageFail)
	}

	item := referenceItem{
		PK:          pk,
		SK:          sk,
		Descriptors: string(data),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal reference item: %w", storage.ErrStorageFail)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName(tableNameReferences)),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to save reference: %w", storage.ErrStorageFail)
	}

	return nil
}

func (s *Storage) getReferencesInternal(ctx context.Context, pk, sk string) ([]manifest.Descriptor, error) {
	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName(tableNameReferences)),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", storage.ErrStorageFail)
	}

	if output.Item == nil {
		return nil, storage.ErrNotFound
	}

	var item referenceItem
	if err := attributevalue.UnmarshalMap(output.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reference: %w", storage.ErrStorageFail)
	}

	var descriptors []manifest.Descriptor
	if err := json.Unmarshal([]byte(item.Descriptors), &descriptors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal descriptors: %w", storage.ErrStorageFail)
	}

	return descriptors, nil
}

func (s *Storage) GetReferences(repoName string, d digest.Digest, artifactType string) ([]manifest.Descriptor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pk := fmt.Sprintf("REPO#%s", repoName)
	sk := fmt.Sprintf("REF#%s", d.String())

	list, err := s.getReferencesInternal(ctx, pk, sk)
	if err != nil {
		return nil, err
	}

	if artifactType == "" {
		return list, nil
	}

	filtered := make([]manifest.Descriptor, 0)
	for _, desc := range list {
		if desc.ArtifactType != nil && *desc.ArtifactType == artifactType {
			filtered = append(filtered, desc)
		}
	}

	return filtered, nil
}

type blobItem struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
}

func (s *Storage) AddBlob(repoName string, d digest.Digest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	item := blobItem{
		PK: fmt.Sprintf("REPO#%s", repoName),
		SK: fmt.Sprintf("BLOB#%s", d.String()),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal blob: %w", storage.ErrStorageFail)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName(tableNameBlobs)),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to add blob: %w", storage.ErrStorageFail)
	}

	return nil
}

func (s *Storage) DeleteBlob(repoName string, d digest.Digest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pk := fmt.Sprintf("REPO#%s", repoName)
	sk := fmt.Sprintf("BLOB#%s", d.String())

	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName(tableNameBlobs)),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to check blob: %w", storage.ErrStorageFail)
	}
	if output.Item == nil {
		return storage.ErrNotFound
	}

	_, err = s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName(tableNameBlobs)),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete blob: %w", storage.ErrStorageFail)
	}

	return nil
}

func (s *Storage) IsExistBlob(repoName string, d digest.Digest) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pk := fmt.Sprintf("REPO#%s", repoName)
	sk := fmt.Sprintf("BLOB#%s", d.String())

	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName(tableNameBlobs)),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to check blob: %w", storage.ErrStorageFail)
	}

	return output.Item != nil, nil
}

func (s *Storage) LinkBlob(newRepo string, d digest.Digest) error {
	return s.AddBlob(newRepo, d)
}
