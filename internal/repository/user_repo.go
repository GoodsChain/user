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
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, updates *models.UpdateUserRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetAllUsers(ctx context.Context, filters *models.FilterParams, sort *models.SortParams, pagination *models.PaginationParams) (*models.GetUsersResponse, error)
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

// GetUserByID retrieves a user by their ID
func (r *postgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := "SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = $1"
	
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	return &user, nil
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

// DeleteUser deletes a user from the database
func (r *postgresUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	// First, get the user data before deletion to return it
	var userToDelete models.User
	selectQuery := "SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = $1"
	err := r.db.GetContext(ctx, &userToDelete, selectQuery, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Delete the user
	deleteQuery := "DELETE FROM users WHERE id = $1"
	result, err := r.db.ExecContext(ctx, deleteQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &userToDelete, nil
}

// GetAllUsers retrieves users with filtering, sorting, and pagination
func (r *postgresUserRepository) GetAllUsers(ctx context.Context, filters *models.FilterParams, sort *models.SortParams, pagination *models.PaginationParams) (*models.GetUsersResponse, error) {
	// Build the base query
	baseQuery := "SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users"
	countQuery := "SELECT COUNT(*) FROM users"
	
	// Build WHERE clause and arguments
	whereClause, args := r.buildWhereClause(filters)
	if whereClause != "" {
		baseQuery += " WHERE " + whereClause
		countQuery += " WHERE " + whereClause
	}
	
	// Get total count for pagination
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	
	// Add ORDER BY clause
	orderClause := r.buildOrderClause(sort)
	baseQuery += " " + orderClause
	
	// Add LIMIT and OFFSET for pagination
	baseQuery += " LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)
	args = append(args, pagination.PageSize, pagination.Offset)
	
	// Execute query
	var users []models.User
	err = r.db.SelectContext(ctx, &users, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	
	// Calculate pagination metadata
	totalPages := (total + pagination.PageSize - 1) / pagination.PageSize
	if totalPages == 0 {
		totalPages = 1
	}
	
	paginationMeta := models.PaginationMetadata{
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}
	
	return &models.GetUsersResponse{
		Data:       users,
		Pagination: paginationMeta,
	}, nil
}

// buildWhereClause constructs the WHERE clause and returns the clause and arguments
func (r *postgresUserRepository) buildWhereClause(filters *models.FilterParams) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argCount := 0
	
	if filters == nil {
		return "", args
	}
	
	if filters.Role != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("role = $%d", argCount))
		args = append(args, *filters.Role)
	}
	
	if filters.IsActive != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, *filters.IsActive)
	}
	
	if filters.Search != nil && *filters.Search != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("(LOWER(full_name) LIKE LOWER($%d) OR LOWER(email) LIKE LOWER($%d))", argCount, argCount))
		args = append(args, "%"+*filters.Search+"%")
	}
	
	if filters.EmailDomain != nil && *filters.EmailDomain != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("email LIKE $%d", argCount))
		args = append(args, "%@"+*filters.EmailDomain+"%")
	}
	
	if filters.CreatedFrom != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argCount))
		args = append(args, *filters.CreatedFrom)
	}
	
	if filters.CreatedTo != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argCount))
		args = append(args, *filters.CreatedTo)
	}
	
	if filters.UpdatedFrom != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("updated_at >= $%d", argCount))
		args = append(args, *filters.UpdatedFrom)
	}
	
	if filters.UpdatedTo != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("updated_at <= $%d", argCount))
		args = append(args, *filters.UpdatedTo)
	}
	
	if len(conditions) == 0 {
		return "", args
	}
	
	whereClause := ""
	for i, condition := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += condition
	}
	
	return whereClause, args
}

// buildOrderClause constructs the ORDER BY clause
func (r *postgresUserRepository) buildOrderClause(sort *models.SortParams) string {
	if sort == nil {
		return "ORDER BY created_at ASC" // Default sorting
	}
	
	// Map field names to database column names
	fieldMap := map[string]string{
		"id":         "id",
		"email":      "email",
		"full_name":  "full_name",
		"role":       "role",
		"is_active":  "is_active",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}
	
	field, exists := fieldMap[sort.Field]
	if !exists {
		field = "created_at" // Default field
	}
	
	order := "ASC"
	if sort.Order == "desc" {
		order = "DESC"
	}
	
	return fmt.Sprintf("ORDER BY %s %s", field, order)
}
