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
