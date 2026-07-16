package repository

import (
	"authentication-service/internal/domain"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Users interface {
	Create(context.Context, *domain.User) error
	GetByEmail(context.Context, string) (*domain.User, error)
	Delete(context.Context, string, string) error
	GetByID(context.Context, string) (*domain.User, error)
	Update(context.Context, *domain.User) error
	UpdateVerificationCode(context.Context, string, domain.VerificationCode) error
	Activate(context.Context, string, bool) error
	AddRole(context.Context, string, string) error
	RemoveRole(context.Context, string, string) error
}
type Sessions interface {
	SetSession(ctx context.Context, session domain.Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error)
}

type Models struct {
	Users    Users
	Sessions Sessions
}

func NewModels(db *pgxpool.Pool) *Models {
	return &Models{
		Users:    NewUsersRepo(db),
		Sessions: NewSessionRepo(db),
	}
}
