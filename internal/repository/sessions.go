package repository

import (
	"authentication-service/internal/domain"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgForeignKeyViolation is the SQLSTATE raised when a session references a user
// that no longer exists.
const pgForeignKeyViolation = "23503"

type SessionRepo struct {
	db *pgxpool.Pool
}

func NewSessionRepo(db *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{db: db}
}

// SetSession stores the user's single active session, replacing any existing one
// (refresh tokens rotate on every refresh). The upsert targets the unique index
// on userid.
func (r *SessionRepo) SetSession(ctx context.Context, session domain.Session) error {
	const query = `
		INSERT INTO sessions (userid, refresh_token, expires_at)
		VALUES ($1::uuid, $2, $3)
		ON CONFLICT (userid) DO UPDATE
		SET refresh_token = EXCLUDED.refresh_token, expires_at = EXCLUDED.expires_at`
	_, err := r.db.Exec(ctx, query, session.UserID, session.RefreshToken, session.ExpiredAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *SessionRepo) GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.QueryRow(ctx,
		`SELECT userid::text, refresh_token, expires_at FROM sessions WHERE refresh_token = $1`,
		refreshToken,
	).Scan(&session.UserID, &session.RefreshToken, &session.ExpiredAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &session, nil
}
