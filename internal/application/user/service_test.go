package user

import (
	"context"
	"io"
	"log/slog"
	"testing"

	domain "github.com/bandrosh/boilerplate-api/internal/domain/user"
)

type fakeRepo struct {
	users  map[domain.ID]*domain.User
	emails map[string]domain.ID
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		users:  make(map[domain.ID]*domain.User),
		emails: make(map[string]domain.ID),
	}
}

func (f *fakeRepo) Create(_ context.Context, u *domain.User) error {
	if _, ok := f.emails[u.Email().String()]; ok {
		return domain.ErrAlreadyExists
	}
	f.users[u.ID()] = u
	f.emails[u.Email().String()] = u.ID()
	return nil
}

func (f *fakeRepo) GetByID(_ context.Context, id domain.ID) (*domain.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (f *fakeRepo) List(_ context.Context, limit int32, _ string) (domain.Page, error) {
	out := make([]*domain.User, 0, len(f.users))
	for _, u := range f.users {
		out = append(out, u)
		if int32(len(out)) == limit {
			break
		}
	}
	return domain.Page{Users: out}, nil
}

func (f *fakeRepo) Update(_ context.Context, u *domain.User) error {
	if _, ok := f.users[u.ID()]; !ok {
		return domain.ErrNotFound
	}
	f.users[u.ID()] = u
	return nil
}

func (f *fakeRepo) Delete(_ context.Context, id domain.ID) error {
	u, ok := f.users[id]
	if !ok {
		return domain.ErrNotFound
	}
	delete(f.emails, u.Email().String())
	delete(f.users, id)
	return nil
}

func newService() *Service {
	return NewService(newFakeRepo(), slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestServiceCreate(t *testing.T) {
	svc := newService()
	ctx := context.Background()

	u, err := svc.Create(ctx, CreateInput{Name: "Ada", Email: "ada@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name() != "Ada" {
		t.Fatalf("got %q", u.Name())
	}

	if _, err := svc.Create(ctx, CreateInput{Name: "X", Email: "bad"}); err != domain.ErrInvalidEmail {
		t.Fatalf("got %v, want ErrInvalidEmail", err)
	}

	if _, err := svc.Create(ctx, CreateInput{Name: "Dup", Email: "ada@example.com"}); err != domain.ErrAlreadyExists {
		t.Fatalf("got %v, want ErrAlreadyExists", err)
	}
}

func TestServiceGetUpdateDelete(t *testing.T) {
	svc := newService()
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateInput{Name: "Ada", Email: "ada@example.com"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := svc.Get(ctx, created.ID())
	if err != nil || got.ID() != created.ID() {
		t.Fatalf("get failed: %v", err)
	}

	updated, err := svc.Update(ctx, UpdateInput{ID: created.ID(), Name: "Ada King"})
	if err != nil || updated.Name() != "Ada King" {
		t.Fatalf("update failed: %v name=%v", err, updated.Name())
	}

	if err := svc.Delete(ctx, created.ID()); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.Get(ctx, created.ID()); err != domain.ErrNotFound {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestServiceListClampsLimit(t *testing.T) {
	svc := newService()
	ctx := context.Background()

	page, err := svc.List(ctx, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Users == nil {
		t.Fatal("expected non-nil slice")
	}
}
