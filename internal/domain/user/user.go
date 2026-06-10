package user

import (
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ID = uuid.UUID

func ParseID(s string) (ID, error) { return uuid.Parse(s) }

type Email struct {
	value string
}

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

func (e Email) String() string { return e.value }

type User struct {
	id        ID
	name      string
	email     Email
	createdAt time.Time
	updatedAt time.Time
}

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

func Hydrate(id ID, name string, email Email, createdAt, updatedAt time.Time) *User {
	return &User{
		id:        id,
		name:      name,
		email:     email,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (u *User) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrNameRequired
	}
	u.name = name
	u.updatedAt = time.Now().UTC()
	return nil
}

func (u *User) ID() ID               { return u.id }
func (u *User) Name() string         { return u.name }
func (u *User) Email() Email         { return u.email }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }
