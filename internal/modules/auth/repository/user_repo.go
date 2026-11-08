package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/auth/model"
	"vendorsaathi/internal/platform/database"
	"vendorsaathi/internal/shared"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) db(ctx context.Context) database.DBTX {
	return database.DBFromContext(ctx, r.pool)
}

func (r *UserRepository) CreateUser(ctx context.Context, phone, email, passwordHash string) (*model.User, error) {
	return r.CreateUserWithStatus(ctx, phone, email, passwordHash, model.UserStatusActive)
}

func (r *UserRepository) CreateUserWithStatus(ctx context.Context, phone, email, passwordHash string, status model.UserStatus) (*model.User, error) {
	var user model.User
	err := r.db(ctx).QueryRow(ctx, `
		INSERT INTO users (phone, email, password_hash, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, phone, email, password_hash, status, preferred_language, created_at, updated_at
	`, phone, nullString(email), passwordHash, status).Scan(
		&user.ID, &user.Phone, &user.Email, &user.PasswordHash, &user.Status, &user.PreferredLanguage, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, shared.NewConflict("phone or email already registered")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db(ctx).QueryRow(ctx, `
		SELECT id, phone, email, password_hash, status, preferred_language, created_at, updated_at
		FROM users WHERE phone = $1
	`, phone).Scan(&user.ID, &user.Phone, &user.Email, &user.PasswordHash, &user.Status, &user.PreferredLanguage, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db(ctx).QueryRow(ctx, `
		SELECT id, phone, email, password_hash, status, preferred_language, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Phone, &user.Email, &user.PasswordHash, &user.Status, &user.PreferredLanguage, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db(ctx).QueryRow(ctx, `
		SELECT id, phone, email, password_hash, status, preferred_language, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Phone, &user.Email, &user.PasswordHash, &user.Status, &user.PreferredLanguage, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.db(ctx).Query(ctx, `
		SELECT r.code FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		roles = append(roles, code)
	}
	return roles, rows.Err()
}

func (r *UserRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.db(ctx).Query(ctx, `
		SELECT DISTINCT p.code FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		permissions = append(permissions, code)
	}
	return permissions, rows.Err()
}

func (r *UserRepository) UpdateUserStatus(ctx context.Context, id uuid.UUID, status model.UserStatus) error {
	cmdTag, err := r.db(ctx).Exec(ctx, `
		UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2
	`, status, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	cmdTag, err := r.db(ctx).Exec(ctx, `
		UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2
	`, passwordHash, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *UserRepository) CreateRefreshSession(ctx context.Context, session *model.RefreshSession) error {
	_, err := r.db(ctx).Exec(ctx, `
		INSERT INTO refresh_sessions (id, user_id, token_hash, expires_at, device_id)
		VALUES ($1, $2, $3, $4, $5)
	`, session.ID, session.UserID, session.TokenHash, session.ExpiresAt, session.DeviceID)
	return err
}

func (r *UserRepository) GetRefreshSession(ctx context.Context, tokenHash string) (*model.RefreshSession, error) {
	var session model.RefreshSession
	err := r.db(ctx).QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, device_id, created_at
		FROM refresh_sessions WHERE token_hash = $1
	`, tokenHash).Scan(&session.ID, &session.UserID, &session.TokenHash, &session.ExpiresAt, &session.RevokedAt, &session.DeviceID, &session.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (r *UserRepository) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	_, err := r.db(ctx).Exec(ctx, `
		UPDATE refresh_sessions SET revoked_at = NOW() WHERE token_hash = $1
	`, tokenHash)
	return err
}

func (r *UserRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db(ctx).Exec(ctx, `
		UPDATE refresh_sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}

func (r *UserRepository) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := r.db(ctx).Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, roleID)
	return err
}

func (r *UserRepository) GetRoleByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	err := r.db(ctx).QueryRow(ctx, `
		SELECT id, code, description, created_at FROM roles WHERE code = $1
	`, code).Scan(&role.ID, &role.Code, &role.Description, &role.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
