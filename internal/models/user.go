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

// FilterParams represents the filtering parameters for user queries
type FilterParams struct {
	Role         *string    `json:"role,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
	Search       *string    `json:"search,omitempty"`
	EmailDomain  *string    `json:"email_domain,omitempty"`
	CreatedFrom  *time.Time `json:"created_from,omitempty"`
	CreatedTo    *time.Time `json:"created_to,omitempty"`
	UpdatedFrom  *time.Time `json:"updated_from,omitempty"`
	UpdatedTo    *time.Time `json:"updated_to,omitempty"`
}

// SortParams represents the sorting parameters for user queries
type SortParams struct {
	Field string `json:"field" validate:"required,oneof=id email full_name role is_active created_at updated_at"`
	Order string `json:"order" validate:"required,oneof=asc desc"`
}

// PaginationParams represents the pagination parameters for user queries
type PaginationParams struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
	Offset   int `json:"-"` // Calculated field, not from request
}

// PaginationMetadata represents the pagination information in the response
type PaginationMetadata struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// GetUsersResponse represents the response for getting multiple users
type GetUsersResponse struct {
	Data       []User             `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

// GetUsersRequest represents the request parameters for getting users
type GetUsersRequest struct {
	// Filtering
	Role        *string `form:"role" validate:"omitempty,oneof=admin staff supplier"`
	IsActive    *bool   `form:"is_active"`
	Search      *string `form:"search"`
	EmailDomain *string `form:"email_domain"`
	CreatedFrom *string `form:"created_from"` // Will be parsed to time.Time
	CreatedTo   *string `form:"created_to"`   // Will be parsed to time.Time
	UpdatedFrom *string `form:"updated_from"` // Will be parsed to time.Time
	UpdatedTo   *string `form:"updated_to"`   // Will be parsed to time.Time
	
	// Sorting
	SortBy    *string `form:"sort_by" validate:"omitempty,oneof=id email full_name role is_active created_at updated_at"`
	SortOrder *string `form:"sort_order" validate:"omitempty,oneof=asc desc"`
	
	// Pagination
	Page     *int `form:"page" validate:"omitempty,min=1"`
	PageSize *int `form:"page_size" validate:"omitempty,min=1,max=100"`
}
