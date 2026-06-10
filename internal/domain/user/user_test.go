package user

import "testing"

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{"valid", "ada@example.com", "ada@example.com", nil},
		{"normalizes case and spaces", "  ADA@Example.COM ", "ada@example.com", nil},
		{"empty", "", "", ErrEmailRequired},
		{"invalid", "not-an-email", "", ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := NewEmail(tt.in)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("got err %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if e.String() != tt.want {
				t.Fatalf("got %q, want %q", e.String(), tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	email, _ := NewEmail("ada@example.com")

	u, err := New("Ada Lovelace", email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID().String() == "" {
		t.Fatal("expected an id")
	}
	if u.Name() != "Ada Lovelace" {
		t.Fatalf("got name %q", u.Name())
	}
	if u.CreatedAt().IsZero() || u.UpdatedAt().IsZero() {
		t.Fatal("expected timestamps to be set")
	}

	if _, err := New("   ", email); err != ErrNameRequired {
		t.Fatalf("got %v, want ErrNameRequired", err)
	}
}

func TestUserRename(t *testing.T) {
	email, _ := NewEmail("ada@example.com")
	u, _ := New("Ada", email)
	created := u.UpdatedAt()

	if err := u.Rename("Ada King"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name() != "Ada King" {
		t.Fatalf("got %q", u.Name())
	}
	if u.UpdatedAt().Before(created) {
		t.Fatal("expected updatedAt to advance")
	}

	if err := u.Rename(""); err != ErrNameRequired {
		t.Fatalf("got %v, want ErrNameRequired", err)
	}
}
