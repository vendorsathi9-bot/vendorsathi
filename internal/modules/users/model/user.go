package model

import (
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

type UserWithRoles struct {
	User  User
	Roles []Role
}

type UpdateProfileInput struct {
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

type UpdateStatusInput struct {
	Status UserStatus `json:"status" binding:"required,oneof=PENDING ACTIVE SUSPENDED CLOSED"`
}

type AssignRoleInput struct {
	RoleCode string `json:"role_code" binding:"required"`
}

type RevokeRoleInput struct {
	RoleCode string `json:"role_code" binding:"required"`
}

type UserOutput struct {
	ID        uuid.UUID  `json:"id"`
	Phone     string     `json:"phone"`
	Email     *string    `json:"email,omitempty"`
	Status    UserStatus `json:"status"`
	Roles     []string   `json:"roles"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type UserListOutput struct {
	Users      []UserOutput `json:"data"`
	NextCursor string       `json:"next_cursor,omitempty"`
	HasMore    bool         `json:"has_more"`
}

func (u *User) ToOutput(roles []string) UserOutput {
	return UserOutput{
		ID:        u.ID,
		Phone:     u.Phone,
		Email:     u.Email,
		Status:    u.Status,
		Roles:     roles,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
