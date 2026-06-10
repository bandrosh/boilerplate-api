package user

import "context"

type Page struct {
	Users      []*User
	NextCursor string
}

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id ID) (*User, error)
	List(ctx context.Context, limit int32, cursor string) (Page, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id ID) error
}
