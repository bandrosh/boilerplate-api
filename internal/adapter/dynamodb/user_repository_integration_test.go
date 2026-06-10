//go:build integration

package dynamodb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/bandrosh/boilerplate-api/internal/platform/config"

	domain "github.com/bandrosh/boilerplate-api/internal/domain/user"
)

func setupTable(t *testing.T) (*UserRepository, func()) {
	t.Helper()
	ctx := context.Background()

	cfg := config.AWS{
		Region:          "us-east-1",
		Endpoint:        "http://127.0.0.1:4566",
		AccessKeyID:     "test",
		SecretAccessKey: "test",
		DynamoTable:     fmt.Sprintf("boilerplate_it_%d", time.Now().UnixNano()),
	}

	client, err := NewClient(ctx, cfg)
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	attr := func(n string) ddbtypes.AttributeDefinition {
		return ddbtypes.AttributeDefinition{AttributeName: aws.String(n), AttributeType: ddbtypes.ScalarAttributeTypeS}
	}
	_, err = client.CreateTable(ctx, &awsdynamodb.CreateTableInput{
		TableName:   aws.String(cfg.DynamoTable),
		BillingMode: ddbtypes.BillingModePayPerRequest,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			attr("PK"), attr("SK"), attr("GSI1PK"), attr("GSI1SK"),
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: ddbtypes.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: ddbtypes.KeyTypeRange},
		},
		GlobalSecondaryIndexes: []ddbtypes.GlobalSecondaryIndex{{
			IndexName: aws.String(gsi1Name),
			KeySchema: []ddbtypes.KeySchemaElement{
				{AttributeName: aws.String("GSI1PK"), KeyType: ddbtypes.KeyTypeHash},
				{AttributeName: aws.String("GSI1SK"), KeyType: ddbtypes.KeyTypeRange},
			},
			Projection: &ddbtypes.Projection{ProjectionType: ddbtypes.ProjectionTypeAll},
		}},
	})
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	waiter := awsdynamodb.NewTableExistsWaiter(client)
	if err := waiter.Wait(ctx, &awsdynamodb.DescribeTableInput{TableName: aws.String(cfg.DynamoTable)}, 30*time.Second); err != nil {
		t.Fatalf("wait table: %v", err)
	}

	cleanup := func() {
		_, _ = client.DeleteTable(context.Background(), &awsdynamodb.DeleteTableInput{TableName: aws.String(cfg.DynamoTable)})
	}
	return NewUserRepository(client, cfg.DynamoTable), cleanup
}

func mustUser(t *testing.T, name, email string) *domain.User {
	t.Helper()
	e, err := domain.NewEmail(email)
	if err != nil {
		t.Fatalf("email: %v", err)
	}
	u, err := domain.New(name, e)
	if err != nil {
		t.Fatalf("new user: %v", err)
	}
	return u
}

func TestUserRepository_CRUD(t *testing.T) {
	repo, cleanup := setupTable(t)
	defer cleanup()
	ctx := context.Background()

	u := mustUser(t, "Ada Lovelace", "ada@example.com")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, u.ID())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Email().String() != "ada@example.com" {
		t.Fatalf("got email %q", got.Email().String())
	}

	dup := mustUser(t, "Other", "ada@example.com")
	if err := repo.Create(ctx, dup); err != domain.ErrAlreadyExists {
		t.Fatalf("got %v, want ErrAlreadyExists", err)
	}

	if err := got.Rename("Ada King"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}

	page, err := repo.List(ctx, 10, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Users) != 1 || page.Users[0].Name() != "Ada King" {
		t.Fatalf("unexpected list: %+v", page.Users)
	}

	if err := repo.Delete(ctx, u.ID()); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, u.ID()); err != domain.ErrNotFound {
		t.Fatalf("got %v, want ErrNotFound", err)
	}

	reuse := mustUser(t, "Reuse", "ada@example.com")
	if err := repo.Create(ctx, reuse); err != nil {
		t.Fatalf("expected email reusable after delete, got: %v", err)
	}
}
