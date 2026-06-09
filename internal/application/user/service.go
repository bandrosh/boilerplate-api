// Package user (application) holds the use cases for the User aggregate. It
// orchestrates the domain and the repository port. It contains no transport or
// persistence detail — only application logic.
package user

import (
	"context"
	"log/slog"

	domain "github.com/bandrosh/boilerplate-api/internal/domain/user"
)

// Service exposes the User use cases. Handlers (inbound adapters) depend on
// this type; it depends only on the domain repository port.
type Service struct {
	repo domain.Repository
	log  *slog.Logger
}

// NewService wires the use cases with their dependencies.
func NewService(repo domain.Repository, log *slog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// CreateInput is the command for the Create use case.
type CreateInput struct {
	Name  string
	Email string
}

// Create registers a new user, enforcing domain invariants.
func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.User, error) {
	email, err := domain.NewEmail(in.Email)
	if err != nil {
		return nil, err
	}

	u, err := domain.New(in.Name, email)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	s.log.InfoContext(ctx, "user created", slog.String("user_id", u.ID().String()))
	return u, nil
}

// Get returns a user by id.
func (s *Service) Get(ctx context.Context, id domain.ID) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a cursor-paginated page of users.
func (s *Service) List(ctx context.Context, limit int32, cursor string) (domain.Page, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, cursor)
}

// UpdateInput is the command for the Update use case.
type UpdateInput struct {
	ID   domain.ID
	Name string
}

// Update changes a user's mutable data.
func (s *Service) Update(ctx context.Context, in UpdateInput) (*domain.User, error) {
	u, err := s.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if err := u.Rename(in.Name); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Delete removes a user by id.
func (s *Service) Delete(ctx context.Context, id domain.ID) error {
	return s.repo.Delete(ctx, id)
}
