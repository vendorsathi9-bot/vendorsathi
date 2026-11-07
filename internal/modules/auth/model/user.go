package model

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusPending   UserStatus = "PENDING"
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
	UserStatusClosed    UserStatus = "CLOSED"
)

type User struct {
	ID           uuid.UUID  `db:"id"`
	Phone        string     `db:"phone"`
	Email        *string    `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Status       UserStatus `db:"status"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

type RefreshSession struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	DeviceID  *string    `db:"device_id"`
	CreatedAt time.Time  `db:"created_at"`
}

type Role struct {
	ID          uuid.UUID `db:"id"`
	Code        string    `db:"code"`
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

type Permission struct {
	ID          uuid.UUID `db:"id"`
	Code        string    `db:"code"`
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

type UserRole struct {
	UserID uuid.UUID `db:"user_id"`
	RoleID uuid.UUID `db:"role_id"`
}

type RolePermission struct {
	RoleID       uuid.UUID `db:"role_id"`
	PermissionID uuid.UUID `db:"permission_id"`
}

type LoginInput struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterInput struct {
	Phone    string `json:"phone" binding:"required"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password" binding:"required,min=8"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenOutput struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	TokenType    string   `json:"token_type"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID     uuid.UUID  `json:"id"`
	Phone  string     `json:"phone"`
	Email  *string    `json:"email,omitempty"`
	Status UserStatus `json:"status"`
	Roles  []string   `json:"roles"`
}

type MeOutput struct {
	User        UserInfo `json:"user"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

func (u *User) ToUserInfo(roles []string) UserInfo {
	return UserInfo{
		ID:     u.ID,
		Phone:  u.Phone,
		Email:  u.Email,
		Status: u.Status,
		Roles:  roles,
	}
}

type pgUUID uuid.UUID

func (p *pgUUID) Scan(src interface{}) error {
	return (*uuid.UUID)(p).Scan(src)
}

func (p pgUUID) Value() (driver.Value, error) {
	return uuid.UUID(p).Value()
}
