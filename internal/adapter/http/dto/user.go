// Package dto holds the request/response shapes for the HTTP transport. Keeping
// them separate from domain types lets the API evolve independently of the model.
package dto

import (
	"time"

	domain "github.com/bandrosh/boilerplate-api/internal/domain/user"
)

// CreateUserRequest is the body for creating a user.
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateUserRequest is the body for updating a user.
type UpdateUserRequest struct {
	Name string `json:"name"`
}

// UserResponse is the public representation of a user.
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromDomain maps a domain User to its API representation.
func FromDomain(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID().String(),
		Name:      u.Name(),
		Email:     u.Email().String(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}

// FromDomainList maps a slice of domain Users to API representations.
func FromDomainList(users []*domain.User) []UserResponse {
	out := make([]UserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, FromDomain(u))
	}
	return out
}

// PaginatedUsers is the cursor-paginated list response. NextCursor is empty
// when there are no further pages.
type PaginatedUsers struct {
	Items      []UserResponse `json:"items"`
	NextCursor string         `json:"next_cursor"`
}

// FromDomainPage maps a domain Page to its API representation.
func FromDomainPage(p domain.Page) PaginatedUsers {
	return PaginatedUsers{
		Items:      FromDomainList(p.Users),
		NextCursor: p.NextCursor,
	}
}
