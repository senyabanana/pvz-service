package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleClient    UserRole = "client"
	RoleEmployee  UserRole = "employee"
	RoleModerator UserRole = "moderator"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"`
	Role      UserRole  `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func IsValidUserRole(role UserRole) bool {
	switch role {
	case RoleClient, RoleEmployee, RoleModerator:
		return true
	default:
		return false
	}
}
