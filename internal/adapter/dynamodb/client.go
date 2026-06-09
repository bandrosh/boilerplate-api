// Package dynamodb is the outbound persistence adapter backed by Amazon
// DynamoDB (LocalStack in development). It owns the SDK client and the
// repository implementations that satisfy the domain ports.
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

// NewClient builds a DynamoDB client. When AWS_ENDPOINT_URL is set (LocalStack)
// it overrides the base endpoint and uses static credentials; in real AWS the
// endpoint is empty and the default credential chain is used.
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

// HealthChecker verifies DynamoDB connectivity for the readiness probe.
type HealthChecker struct {
	client *dynamodb.Client
	table  string
}

// NewHealthChecker builds the readiness checker for a given table.
func NewHealthChecker(client *dynamodb.Client, table string) *HealthChecker {
	return &HealthChecker{client: client, table: table}
}

// Ping confirms the table is reachable by describing it.
func (h *HealthChecker) Ping(ctx context.Context) error {
	_, err := h.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(h.table),
	})
	return err
}
