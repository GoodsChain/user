package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents the user model in the database
type User struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Email     string     `json:"email" db:"email" validate:"required,email"`
	FullName  string     `json:"full_name" db:"full_name" validate:"required"`
	Phone     *string    `json:"phone" db:"phone"` // Use pointer for nullable fields
	Role      string     `json:"role" db:"role" validate:"required,oneof=admin staff supplier"`
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest represents the request body for creating a new user
type CreateUserRequest struct {
	Email    string  `json:"email" validate:"required,email"`
	FullName string  `json:"full_name" validate:"required"`
	Phone    *string `json:"phone"`
	Role     string  `json:"role" validate:"required,oneof=admin staff supplier"`
}

// UpdateUserRequest represents the request body for updating an existing user
type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	FullName *string `json:"full_name,omitempty" validate:"omitempty,min=1"`
	Phone    *string `json:"phone,omitempty"`
	Role     *string `json:"role,omitempty" validate:"omitempty,oneof=admin staff supplier"`
	IsActive *bool   `json:"is_active,omitempty"`
}
