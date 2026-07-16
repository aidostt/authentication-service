package repository

import (
	"authentication-service/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgUniqueViolation is the SQLSTATE raised when an insert or update conflicts
// with a unique constraint (here: a second user with an existing email). It is
// the Postgres counterpart of MongoDB's duplicate-key error 11000.
const pgUniqueViolation = "23505"

type UsersRepo struct {
	db *pgxpool.Pool
}

func NewUsersRepo(db *pgxpool.Pool) *UsersRepo {
	return &UsersRepo{db: db}
}

// userColumns is the ordered projection scanned by scanUser. id is rendered as
// text so it maps onto the string identifier the wire contract uses, and the
// nullable name/surname/phone default to empty strings.
const userColumns = `id::text, email, password,
	COALESCE(name, ''), COALESCE(surname, ''), COALESCE(phone, ''),
	roles, activated, verification_code, verification_expiry`

func scanUser(row pgx.Row) (*domain.User, error) {
	var (
		u      domain.User
		code   *string
		expiry *time.Time
	)
	if err := row.Scan(
		&u.ID, &u.Email, &u.Password,
		&u.Name, &u.Surname, &u.Phone,
		&u.Roles, &u.Activated, &code, &expiry,
	); err != nil {
		return nil, err
	}
	if code != nil {
		u.VerificationCode.Code = *code
	}
	if expiry != nil {
		u.VerificationCode.ExpiredAt = *expiry
	}
	return &u, nil
}

func (r *UsersRepo) Create(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (email, password, name, surname, phone, roles, activated, verification_code, verification_expiry)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id::text`
	err := r.db.QueryRow(ctx, query,
		user.Email, user.Password, user.Name, user.Surname, user.Phone,
		user.Roles, user.Activated, user.VerificationCode.Code, user.VerificationCode.ExpiredAt,
	).Scan(&user.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UsersRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1::uuid`, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UsersRepo) Delete(ctx context.Context, id, email string) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1::uuid AND email = $2`, id, email)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UsersRepo) Update(ctx context.Context, user *domain.User) error {
	const query = `
		UPDATE users
		SET name = $1, surname = $2, phone = $3, email = $4, roles = $5,
		    password = $6, activated = $7, verification_code = $8, verification_expiry = $9
		WHERE id = $10::uuid`
	ct, err := r.db.Exec(ctx, query,
		user.Name, user.Surname, user.Phone, user.Email, user.Roles,
		user.Password, user.Activated, user.VerificationCode.Code, user.VerificationCode.ExpiredAt,
		user.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrUserAlreadyExists
		}
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UsersRepo) UpdateVerificationCode(ctx context.Context, id string, code domain.VerificationCode) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE users SET verification_code = $1, verification_expiry = $2 WHERE id = $3::uuid`,
		code.Code, code.ExpiredAt, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UsersRepo) Activate(ctx context.Context, id string, activate bool) error {
	ct, err := r.db.Exec(ctx, `UPDATE users SET activated = $1 WHERE id = $2::uuid`, activate, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// AddRole appends role only when it is not already present, preserving the
// existing order. This mirrors MongoDB's $addToSet and, unlike a naive
// array_append, keeps the operation idempotent.
func (r *UsersRepo) AddRole(ctx context.Context, id, role string) error {
	const query = `
		UPDATE users
		SET roles = CASE WHEN $1 = ANY(roles) THEN roles ELSE array_append(roles, $1) END
		WHERE id = $2::uuid`
	ct, err := r.db.Exec(ctx, query, role, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UsersRepo) RemoveRole(ctx context.Context, id, role string) error {
	ct, err := r.db.Exec(ctx, `UPDATE users SET roles = array_remove(roles, $1) WHERE id = $2::uuid`, role, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
