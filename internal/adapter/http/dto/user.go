package dto

import (
	"time"

	domain "github.com/bandrosh/boilerplate-api/internal/domain/user"
)

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func FromDomain(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID().String(),
		Name:      u.Name(),
		Email:     u.Email().String(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}

func FromDomainList(users []*domain.User) []UserResponse {
	out := make([]UserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, FromDomain(u))
	}
	return out
}

type PaginatedUsers struct {
	Items      []UserResponse `json:"items"`
	NextCursor string         `json:"next_cursor"`
}

func FromDomainPage(p domain.Page) PaginatedUsers {
	return PaginatedUsers{
		Items:      FromDomainList(p.Users),
		NextCursor: p.NextCursor,
	}
}
