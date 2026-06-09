package user

import "context"

// Page is a cursor-paginated result set. NextCursor is empty when there are no
// more items. Cursor-based pagination is the idiomatic choice for DynamoDB
// (it has no natural OFFSET).
type Page struct {
	Users      []*User
	NextCursor string
}

// Repository is the outbound port for persisting and retrieving Users. The
// domain defines the interface; a concrete adapter (e.g. DynamoDB) implements
// it. This is the dependency-inversion seam of the hexagonal architecture.
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id ID) (*User, error)
	List(ctx context.Context, limit int32, cursor string) (Page, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id ID) error
}
