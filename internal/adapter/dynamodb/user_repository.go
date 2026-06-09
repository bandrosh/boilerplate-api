package dynamodb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	domain "github.com/bandrosh/boilerplate-api/internal/domain/user"
)

// Single-table design:
//
//	User item       PK = USER#<id>     SK = USER#<id>
//	                GSI1PK = USER       GSI1SK = <createdAt>#<id>   (for listing)
//	Email-lock item PK = EMAIL#<email> SK = EMAIL#<email>          (uniqueness)
//
// Creation writes both items in a single transaction with attribute_not_exists
// conditions, guaranteeing e-mail uniqueness.
const (
	gsi1Name      = "GSI1"
	userPartition = "USER"
)

// UserRepository implements domain/user.Repository on top of DynamoDB.
type UserRepository struct {
	client *dynamodb.Client
	table  string
}

// NewUserRepository builds the repository for a given table.
func NewUserRepository(client *dynamodb.Client, table string) *UserRepository {
	return &UserRepository{client: client, table: table}
}

var _ domain.Repository = (*UserRepository)(nil)

// userItem is the persisted representation of a User (the main item).
type userItem struct {
	PK         string    `dynamodbav:"PK"`
	SK         string    `dynamodbav:"SK"`
	GSI1PK     string    `dynamodbav:"GSI1PK"`
	GSI1SK     string    `dynamodbav:"GSI1SK"`
	EntityType string    `dynamodbav:"EntityType"`
	ID         string    `dynamodbav:"Id"`
	Name       string    `dynamodbav:"Name"`
	Email      string    `dynamodbav:"Email"`
	CreatedAt  time.Time `dynamodbav:"CreatedAt"`
	UpdatedAt  time.Time `dynamodbav:"UpdatedAt"`
}

func userPK(id string) string     { return "USER#" + id }
func emailPK(email string) string { return "EMAIL#" + email }

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	item := userItem{
		PK:         userPK(u.ID().String()),
		SK:         userPK(u.ID().String()),
		GSI1PK:     userPartition,
		GSI1SK:     fmt.Sprintf("%s#%s", u.CreatedAt().UTC().Format(time.RFC3339Nano), u.ID()),
		EntityType: "USER",
		ID:         u.ID().String(),
		Name:       u.Name(),
		Email:      u.Email().String(),
		CreatedAt:  u.CreatedAt(),
		UpdatedAt:  u.UpdatedAt(),
	}

	userAV, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("marshal user: %w", err)
	}

	lockAV, err := attributevalue.MarshalMap(map[string]any{
		"PK":         emailPK(u.Email().String()),
		"SK":         emailPK(u.Email().String()),
		"EntityType": "EMAIL_LOCK",
		"UserId":     u.ID().String(),
	})
	if err != nil {
		return fmt.Errorf("marshal email lock: %w", err)
	}

	_, err = r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []ddbtypes.TransactWriteItem{
			{Put: &ddbtypes.Put{
				TableName:           aws.String(r.table),
				Item:                userAV,
				ConditionExpression: aws.String("attribute_not_exists(PK)"),
			}},
			{Put: &ddbtypes.Put{
				TableName:           aws.String(r.table),
				Item:                lockAV,
				ConditionExpression: aws.String("attribute_not_exists(PK)"),
			}},
		},
	})
	if isTransactionConflict(err) {
		return domain.ErrAlreadyExists
	}
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id domain.ID) (*domain.User, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: userPK(id.String())},
			"SK": &ddbtypes.AttributeValueMemberS{Value: userPK(id.String())},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, domain.ErrNotFound
	}

	var item userItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return toDomain(item)
}

func (r *UserRepository) List(ctx context.Context, limit int32, cursor string) (domain.Page, error) {
	startKey, err := decodeCursor(cursor)
	if err != nil {
		return domain.Page{}, err
	}

	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(r.table),
		IndexName:                 aws.String(gsi1Name),
		KeyConditionExpression:    aws.String("GSI1PK = :pk"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{":pk": &ddbtypes.AttributeValueMemberS{Value: userPartition}},
		ScanIndexForward:          aws.Bool(false), // newest first
		Limit:                     aws.Int32(limit),
		ExclusiveStartKey:         startKey,
	})
	if err != nil {
		return domain.Page{}, err
	}

	users := make([]*domain.User, 0, len(out.Items))
	for _, raw := range out.Items {
		var item userItem
		if err := attributevalue.UnmarshalMap(raw, &item); err != nil {
			return domain.Page{}, fmt.Errorf("unmarshal user: %w", err)
		}
		u, err := toDomain(item)
		if err != nil {
			return domain.Page{}, err
		}
		users = append(users, u)
	}

	next, err := encodeCursor(out.LastEvaluatedKey)
	if err != nil {
		return domain.Page{}, err
	}
	return domain.Page{Users: users, NextCursor: next}, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.table),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: userPK(u.ID().String())},
			"SK": &ddbtypes.AttributeValueMemberS{Value: userPK(u.ID().String())},
		},
		UpdateExpression:    aws.String("SET #name = :name, UpdatedAt = :updated"),
		ConditionExpression: aws.String("attribute_exists(PK)"),
		ExpressionAttributeNames: map[string]string{
			"#name": "Name", // Name is a DynamoDB reserved word
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":name":    &ddbtypes.AttributeValueMemberS{Value: u.Name()},
			":updated": &ddbtypes.AttributeValueMemberS{Value: u.UpdatedAt().UTC().Format(time.RFC3339Nano)},
		},
	})
	if isConditionalCheckFailed(err) {
		return domain.ErrNotFound
	}
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id domain.ID) error {
	// Need the e-mail to remove the uniqueness lock atomically.
	u, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	_, err = r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []ddbtypes.TransactWriteItem{
			{Delete: &ddbtypes.Delete{
				TableName: aws.String(r.table),
				Key: map[string]ddbtypes.AttributeValue{
					"PK": &ddbtypes.AttributeValueMemberS{Value: userPK(id.String())},
					"SK": &ddbtypes.AttributeValueMemberS{Value: userPK(id.String())},
				},
			}},
			{Delete: &ddbtypes.Delete{
				TableName: aws.String(r.table),
				Key: map[string]ddbtypes.AttributeValue{
					"PK": &ddbtypes.AttributeValueMemberS{Value: emailPK(u.Email().String())},
					"SK": &ddbtypes.AttributeValueMemberS{Value: emailPK(u.Email().String())},
				},
			}},
		},
	})
	return err
}

func toDomain(item userItem) (*domain.User, error) {
	id, err := domain.ParseID(item.ID)
	if err != nil {
		return nil, fmt.Errorf("parse id: %w", err)
	}
	email, err := domain.NewEmail(item.Email)
	if err != nil {
		return nil, err
	}
	return domain.Hydrate(id, item.Name, email, item.CreatedAt, item.UpdatedAt), nil
}

// ─── cursor helpers ──────────────────────────────────────────

type cursorKey struct {
	PK     string `dynamodbav:"PK"`
	SK     string `dynamodbav:"SK"`
	GSI1PK string `dynamodbav:"GSI1PK"`
	GSI1SK string `dynamodbav:"GSI1SK"`
}

func encodeCursor(lastKey map[string]ddbtypes.AttributeValue) (string, error) {
	if len(lastKey) == 0 {
		return "", nil
	}
	var ck cursorKey
	if err := attributevalue.UnmarshalMap(lastKey, &ck); err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}
	raw, err := json.Marshal(ck)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeCursor(cursor string) (map[string]ddbtypes.AttributeValue, error) {
	if cursor == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}
	var ck cursorKey
	if err := json.Unmarshal(raw, &ck); err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}
	key, err := attributevalue.MarshalMap(ck)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}
	return key, nil
}

// ─── error classification ────────────────────────────────────

func isTransactionConflict(err error) bool {
	var tce *ddbtypes.TransactionCanceledException
	return errors.As(err, &tce)
}

func isConditionalCheckFailed(err error) bool {
	var cce *ddbtypes.ConditionalCheckFailedException
	return errors.As(err, &cce)
}
