package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/users/model"
	"vendorsaathi/internal/shared"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, `
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

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, `
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

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, `
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

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, phone, email, preferredLanguage string) (*model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, `
		UPDATE users
		SET phone = COALESCE(NULLIF($1, ''), phone),
		    email = COALESCE(NULLIF($2, ''), email),
		    preferred_language = COALESCE(NULLIF($3, ''), preferred_language),
		    updated_at = NOW()
		WHERE id = $4
		RETURNING id, phone, email, password_hash, status, preferred_language, created_at, updated_at
	`, phone, email, preferredLanguage, id).Scan(&user.ID, &user.Phone, &user.Email, &user.PasswordHash, &user.Status, &user.PreferredLanguage, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, shared.NewConflict("phone or email already in use")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.UserStatus) error {
	cmdTag, err := r.pool.Exec(ctx, `
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
	cmdTag, err := r.pool.Exec(ctx, `
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

func (r *UserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
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
	rows, err := r.pool.Query(ctx, `
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

func (r *UserRepository) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, roleID)
	return err
}

func (r *UserRepository) RevokeRole(ctx context.Context, userID, roleID uuid.UUID) error {
	cmdTag, err := r.pool.Exec(ctx, `
		DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2
	`, userID, roleID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *UserRepository) GetRoleByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	err := r.pool.QueryRow(ctx, `
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

func (r *UserRepository) ListUsers(ctx context.Context, limit int, cursor string, status *model.UserStatus) ([]model.User, string, bool, error) {
	query := `
		SELECT id, phone, email, password_hash, status, preferred_language, created_at, updated_at
		FROM users
	`
	args := []interface{}{}
	argIdx := 1

	if status != nil {
		query += " WHERE status = $" + string(rune(argIdx+'0'))
		args = append(args, *status)
		argIdx++
	}

	query += " ORDER BY created_at DESC LIMIT $" + string(rune(argIdx+'0'))
	args = append(args, limit+1)
	argIdx++

	if cursor != "" {
		cursorID, err := uuid.Parse(cursor)
		if err != nil {
			return nil, "", false, err
		}
		query = query + " AND id < $" + string(rune(argIdx+'0'))
		args = append(args, cursorID)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", false, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Phone, &user.Email, &user.PasswordHash, &user.Status, &user.PreferredLanguage, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, "", false, err
		}
		users = append(users, user)
	}

	hasMore := len(users) > limit
	if hasMore {
		users = users[:limit]
	}

	var nextCursor string
	if hasMore && len(users) > 0 {
		nextCursor = users[len(users)-1].ID.String()
	}

	return users, nextCursor, hasMore, rows.Err()
}
