package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/GoodsChain/user/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, updates *models.UpdateUserRequest) (*models.User, error)
}

// postgresUserRepository implements UserRepository for PostgreSQL
type postgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new instance of postgresUserRepository
func NewPostgresUserRepository(db *sqlx.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

// CreateUser inserts a new user into the database
func (r *postgresUserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true // Default to active

	query := `
		INSERT INTO users (id, email, full_name, phone, role, is_active, created_at, updated_at)
		VALUES (:id, :email, :full_name, :phone, :role, :is_active, :created_at, :updated_at)
		RETURNING id, email, full_name, phone, role, is_active, created_at, updated_at
	`

	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var createdUser models.User
		if err := rows.StructScan(&createdUser); err != nil {
			return nil, fmt.Errorf("failed to scan created user: %w", err)
		}
		return &createdUser, nil
	}

	return nil, fmt.Errorf("no user returned after insert")
}

// UpdateUser updates an existing user in the database
func (r *postgresUserRepository) UpdateUser(ctx context.Context, id uuid.UUID, updates *models.UpdateUserRequest) (*models.User, error) {
	// First, check if the user exists
	var existingUser models.User
	checkQuery := "SELECT id FROM users WHERE id = $1"
	err := r.db.GetContext(ctx, &existingUser, checkQuery, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Build dynamic query based on provided fields
	setParts := []string{}
	args := map[string]interface{}{
		"id":         id,
		"updated_at": time.Now(),
	}

	if updates.Email != nil {
		setParts = append(setParts, "email = :email")
		args["email"] = *updates.Email
	}
	if updates.FullName != nil {
		setParts = append(setParts, "full_name = :full_name")
		args["full_name"] = *updates.FullName
	}
	if updates.Phone != nil {
		setParts = append(setParts, "phone = :phone")
		args["phone"] = *updates.Phone
	}
	if updates.Role != nil {
		setParts = append(setParts, "role = :role")
		args["role"] = *updates.Role
	}
	if updates.IsActive != nil {
		setParts = append(setParts, "is_active = :is_active")
		args["is_active"] = *updates.IsActive
	}

	// Always update the updated_at field
	setParts = append(setParts, "updated_at = :updated_at")

	if len(setParts) == 1 { // Only updated_at was added
		return nil, fmt.Errorf("no fields to update")
	}

	// Build the SET clause
	setClause := ""
	for i, part := range setParts {
		if i > 0 {
			setClause += ", "
		}
		setClause += part
	}

	query := fmt.Sprintf(`
		UPDATE users 
		SET %s 
		WHERE id = :id
		RETURNING id, email, full_name, phone, role, is_active, created_at, updated_at
	`, setClause)

	rows, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var updatedUser models.User
		if err := rows.StructScan(&updatedUser); err != nil {
			return nil, fmt.Errorf("failed to scan updated user: %w", err)
		}
		return &updatedUser, nil
	}

	return nil, fmt.Errorf("no user returned after update")
}
