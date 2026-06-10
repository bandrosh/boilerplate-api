package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/bandrosh/boilerplate-api/internal/platform/config"
)

func NewClient(ctx context.Context, cfg config.AWS) (*dynamodb.Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.Endpoint != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	}), nil
}

type HealthChecker struct {
	client *dynamodb.Client
	table  string
}

func NewHealthChecker(client *dynamodb.Client, table string) *HealthChecker {
	return &HealthChecker{client: client, table: table}
}

func (h *HealthChecker) Ping(ctx context.Context) error {
	_, err := h.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(h.table),
	})
	return err
}
