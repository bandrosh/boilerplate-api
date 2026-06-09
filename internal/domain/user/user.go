// Package user is the domain core for the User aggregate. It holds entities,
// value objects, domain errors and the repository port (interface). It has no
// dependency on frameworks, transport or persistence — those are adapters.
package user

import (
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ID is the unique identifier of a User (a value object wrapping uuid).
type ID = uuid.UUID

// ParseID parses a string into a User ID.
func ParseID(s string) (ID, error) { return uuid.Parse(s) }

// Email is a validated value object. It can only be created through NewEmail,
// which guarantees the invariant "this is a syntactically valid e-mail".
type Email struct {
	value string
}

// NewEmail validates and builds an Email value object.
func NewEmail(raw string) (Email, error) {
	addr := strings.TrimSpace(strings.ToLower(raw))
	if addr == "" {
		return Email{}, ErrEmailRequired
	}
	if _, err := mail.ParseAddress(addr); err != nil {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: addr}, nil
}

// String returns the underlying e-mail address.
func (e Email) String() string { return e.value }

// User is the aggregate root. Fields are unexported so the entity can only be
// mutated through behaviour that preserves its invariants.
type User struct {
	id        ID
	name      string
	email     Email
	createdAt time.Time
	updatedAt time.Time
}

// New creates a brand-new User, enforcing all invariants. Use this for
// registration; use Hydrate to rebuild a User loaded from persistence.
func New(name string, email Email) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameRequired
	}

	now := time.Now().UTC()
	return &User{
		id:        uuid.New(),
		name:      name,
		email:     email,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Hydrate rebuilds a User from already-persisted data without re-running
// creation rules. Intended for use by repository adapters only.
func Hydrate(id ID, name string, email Email, createdAt, updatedAt time.Time) *User {
	return &User{
		id:        id,
		name:      name,
		email:     email,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Rename changes the user's name, keeping invariants intact.
func (u *User) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrNameRequired
	}
	u.name = name
	u.updatedAt = time.Now().UTC()
	return nil
}

// Accessors — read-only views of the aggregate state.
func (u *User) ID() ID               { return u.id }
func (u *User) Name() string         { return u.name }
func (u *User) Email() Email         { return u.email }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }
