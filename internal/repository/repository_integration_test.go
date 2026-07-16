package repository

import (
	"authentication-service/internal/domain"
	"authentication-service/internal/repository/testsupport"
	"context"
	"errors"
	"testing"
	"time"
)

func setup(t *testing.T) *Models {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}
	pool, cleanup, err := testsupport.SetupPostgres(context.Background())
	if err != nil {
		t.Skipf("cannot start postgres (docker unavailable?): %v", err)
	}
	t.Cleanup(cleanup)
	return NewModels(pool)
}

func newUser(email string) *domain.User {
	return &domain.User{
		Name:     "Ada",
		Surname:  "Lovelace",
		Phone:    "+10000000000",
		Email:    email,
		Roles:    []string{domain.UserRole},
		Password: "hashed-password",
		VerificationCode: domain.VerificationCode{
			Code:      "123456",
			ExpiredAt: time.Now().Add(time.Hour).UTC(),
		},
	}
}

func TestUsersRepo_Lifecycle(t *testing.T) {
	ctx := context.Background()
	m := setup(t)

	user := newUser("ada@example.com")
	if err := m.Users.Create(ctx, user); err != nil {
		t.Fatalf("create: %v", err)
	}
	if user.ID == "" {
		t.Fatal("Create did not populate the generated id")
	}

	// A second user with the same email must surface as a unique violation
	// mapped to ErrUserAlreadyExists (SQLSTATE 23505).
	dup := newUser("ada@example.com")
	if err := m.Users.Create(ctx, dup); !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("duplicate email: got %v, want ErrUserAlreadyExists", err)
	}

	byEmail, err := m.Users.GetByEmail(ctx, "ada@example.com")
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if byEmail.ID != user.ID || byEmail.Name != "Ada" {
		t.Fatalf("get by email mismatch: %+v", byEmail)
	}
	if byEmail.VerificationCode.Code != "123456" {
		t.Fatalf("verification code not persisted: got %q", byEmail.VerificationCode.Code)
	}

	byID, err := m.Users.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if byID.Email != "ada@example.com" {
		t.Fatalf("get by id mismatch: %+v", byID)
	}

	if err := m.Users.Activate(ctx, user.ID, true); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := m.Users.AddRole(ctx, user.ID, domain.ActivatedRole); err != nil {
		t.Fatalf("add role: %v", err)
	}
	// AddRole is idempotent: adding the same role again must not duplicate it or
	// error out.
	if err := m.Users.AddRole(ctx, user.ID, domain.ActivatedRole); err != nil {
		t.Fatalf("add role (idempotent): %v", err)
	}

	got, err := m.Users.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get after activate: %v", err)
	}
	if !got.Activated {
		t.Fatal("expected user to be activated")
	}
	if n := countRole(got.Roles, domain.ActivatedRole); n != 1 {
		t.Fatalf("expected exactly one %q role, got %d in %v", domain.ActivatedRole, n, got.Roles)
	}

	if _, err := m.Users.GetByID(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("missing user: got %v, want ErrUserNotFound", err)
	}
}

func TestSessionsRepo_UpsertAndLookup(t *testing.T) {
	ctx := context.Background()
	m := setup(t)

	user := newUser("grace@example.com")
	if err := m.Users.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	exp := time.Now().Add(12 * time.Hour).UTC()
	if err := m.Sessions.SetSession(ctx, domain.Session{UserID: user.ID, RefreshToken: "rt-1", ExpiredAt: exp}); err != nil {
		t.Fatalf("set session: %v", err)
	}
	// SetSession upserts on userid, so a second call replaces the first token.
	if err := m.Sessions.SetSession(ctx, domain.Session{UserID: user.ID, RefreshToken: "rt-2", ExpiredAt: exp}); err != nil {
		t.Fatalf("upsert session: %v", err)
	}

	if _, err := m.Sessions.GetByRefreshToken(ctx, "rt-1"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("stale token should be gone: got %v, want ErrUserNotFound", err)
	}

	session, err := m.Sessions.GetByRefreshToken(ctx, "rt-2")
	if err != nil {
		t.Fatalf("get by refresh token: %v", err)
	}
	if session.UserID != user.ID {
		t.Fatalf("session user mismatch: got %s, want %s", session.UserID, user.ID)
	}

	// Deleting the user cascades to the session (FK ON DELETE CASCADE).
	if err := m.Users.Delete(ctx, user.ID, user.Email); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	if _, err := m.Sessions.GetByRefreshToken(ctx, "rt-2"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("session should cascade-delete: got %v, want ErrUserNotFound", err)
	}
}

func countRole(roles []string, target string) int {
	n := 0
	for _, r := range roles {
		if r == target {
			n++
		}
	}
	return n
}
