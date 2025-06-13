package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GoodsChain/user/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateUser_Success tests successful user creation with all fields
func TestCreateUser_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Test data
	phone := "1234567890"
	inputUser := &models.User{
		Email:    "test@example.com",
		FullName: "John Doe",
		Phone:    &phone,
		Role:     "admin",
	}

	expectedID := uuid.New()
	expectedTime := time.Now()

	// Define the columns that will be returned
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}

	// Set up mock expectation
	mock.ExpectQuery(`INSERT INTO users \(id, email, full_name, phone, role, is_active, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8\) RETURNING id, email, full_name, phone, role, is_active, created_at, updated_at`).
		WithArgs(sqlmock.AnyArg(), "test@example.com", "John Doe", &phone, "admin", true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(expectedID, "test@example.com", "John Doe", &phone, "admin", true, expectedTime, expectedTime))

	// Execute the method
	ctx := context.Background()
	result, err := repo.CreateUser(ctx, inputUser)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedID, result.ID)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "John Doe", result.FullName)
	assert.Equal(t, &phone, result.Phone)
	assert.Equal(t, "admin", result.Role)
	assert.True(t, result.IsActive)
	assert.Equal(t, expectedTime, result.CreatedAt)
	assert.Equal(t, expectedTime, result.UpdatedAt)

	// Verify that the input user was modified with generated values
	assert.NotEqual(t, uuid.Nil, inputUser.ID)
	assert.True(t, inputUser.IsActive)
	assert.False(t, inputUser.CreatedAt.IsZero())
	assert.False(t, inputUser.UpdatedAt.IsZero())

	// Verify all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestCreateUser_SuccessMinimalFields tests user creation with minimal required fields
func TestCreateUser_SuccessMinimalFields(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Test data with minimal fields (no phone)
	inputUser := &models.User{
		Email:    "minimal@example.com",
		FullName: "Jane Doe",
		Phone:    nil, // Nullable field
		Role:     "staff",
	}

	expectedID := uuid.New()
	expectedTime := time.Now()

	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}

	mock.ExpectQuery(`INSERT INTO users \(id, email, full_name, phone, role, is_active, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8\) RETURNING id, email, full_name, phone, role, is_active, created_at, updated_at`).
		WithArgs(sqlmock.AnyArg(), "minimal@example.com", "Jane Doe", nil, "staff", true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(expectedID, "minimal@example.com", "Jane Doe", nil, "staff", true, expectedTime, expectedTime))

	ctx := context.Background()
	result, err := repo.CreateUser(ctx, inputUser)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedID, result.ID)
	assert.Equal(t, "minimal@example.com", result.Email)
	assert.Equal(t, "Jane Doe", result.FullName)
	assert.Nil(t, result.Phone)
	assert.Equal(t, "staff", result.Role)
	assert.True(t, result.IsActive)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestCreateUser_DatabaseError tests handling of database execution errors
func TestCreateUser_DatabaseError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	inputUser := &models.User{
		Email:    "error@example.com",
		FullName: "Error User",
		Role:     "admin",
	}

	// Simulate a database error (e.g., unique constraint violation)
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "error@example.com", "Error User", nil, "admin", true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	result, err := repo.CreateUser(ctx, inputUser)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to insert user")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestCreateUser_ScanError tests handling of row scanning errors
func TestCreateUser_ScanError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	inputUser := &models.User{
		Email:    "scan@example.com",
		FullName: "Scan User",
		Role:     "supplier",
	}

	// Return invalid data that will cause scanning to fail
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "scan@example.com", "Scan User", nil, "supplier", true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("invalid-uuid", "scan@example.com", "Scan User", nil, "supplier", true, "invalid-time", "invalid-time"))

	ctx := context.Background()
	result, err := repo.CreateUser(ctx, inputUser)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to scan created user")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestCreateUser_NoRowsReturned tests handling when no rows are returned after insert
func TestCreateUser_NoRowsReturned(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	inputUser := &models.User{
		Email:    "norows@example.com",
		FullName: "No Rows User",
		Role:     "admin",
	}

	// Return empty result set
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "norows@example.com", "No Rows User", nil, "admin", true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(columns)) // Empty rows

	ctx := context.Background()
	result, err := repo.CreateUser(ctx, inputUser)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no user returned after insert")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestCreateUser_ContextCancelled tests handling of context cancellation
func TestCreateUser_ContextCancelled(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	inputUser := &models.User{
		Email:    "context@example.com",
		FullName: "Context User",
		Role:     "staff",
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The query should not be executed due to cancelled context
	// but we still need to set up the expectation in case it does get called
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "context@example.com", "Context User", nil, "staff", true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(context.Canceled)

	result, err := repo.CreateUser(ctx, inputUser)

	assert.Error(t, err)
	assert.Nil(t, result)
	// The error could be context.Canceled or a wrapped version
	assert.True(t, err == context.Canceled || err.Error() == "failed to insert user: context canceled")
}

// TestCreateUser_UniqueConstraintViolation tests handling of unique constraint violations
func TestCreateUser_UniqueConstraintViolation(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	inputUser := &models.User{
		Email:    "duplicate@example.com",
		FullName: "Duplicate User",
		Role:     "admin",
	}

	// Simulate unique constraint violation (email already exists)
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "duplicate@example.com", "Duplicate User", nil, "admin", true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows) // This simulates a constraint violation

	ctx := context.Background()
	result, err := repo.CreateUser(ctx, inputUser)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to insert user")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestCreateUser_AllRoles tests creation with different valid roles
func TestCreateUser_AllRoles(t *testing.T) {
	validRoles := []string{"admin", "staff", "supplier"}

	for _, role := range validRoles {
		t.Run("Role_"+role, func(t *testing.T) {
			db, mock, repo := setupMockDB(t)
			defer db.Close()

			inputUser := &models.User{
				Email:    role + "@example.com",
				FullName: role + " User",
				Role:     role,
			}

			expectedID := uuid.New()
			expectedTime := time.Now()
			columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}

			mock.ExpectQuery(`INSERT INTO users`).
				WithArgs(sqlmock.AnyArg(), role+"@example.com", role+" User", nil, role, true, sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows(columns).
					AddRow(expectedID, role+"@example.com", role+" User", nil, role, true, expectedTime, expectedTime))

			ctx := context.Background()
			result, err := repo.CreateUser(ctx, inputUser)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, role, result.Role)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

// TestCreateUser_VerifyGeneratedFields tests that UUIDs and timestamps are properly generated
func TestCreateUser_VerifyGeneratedFields(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	inputUser := &models.User{
		Email:    "generated@example.com",
		FullName: "Generated User",
		Role:     "admin",
	}

	expectedID := uuid.New()
	expectedTime := time.Now()
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "generated@example.com", "Generated User", nil, "admin", true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(expectedID, "generated@example.com", "Generated User", nil, "admin", true, expectedTime, expectedTime))

	ctx := context.Background()

	// Store original values to verify they were empty
	originalID := inputUser.ID
	originalCreatedAt := inputUser.CreatedAt
	originalUpdatedAt := inputUser.UpdatedAt
	originalIsActive := inputUser.IsActive

	result, err := repo.CreateUser(ctx, inputUser)

	require.NoError(t, err)

	// Verify original user was empty/default
	assert.Equal(t, uuid.Nil, originalID)
	assert.True(t, originalCreatedAt.IsZero())
	assert.True(t, originalUpdatedAt.IsZero())
	assert.False(t, originalIsActive)

	// Verify input user was modified with generated values
	assert.NotEqual(t, uuid.Nil, inputUser.ID)
	assert.False(t, inputUser.CreatedAt.IsZero())
	assert.False(t, inputUser.UpdatedAt.IsZero())
	assert.True(t, inputUser.IsActive)

	// Verify returned user has expected values
	assert.Equal(t, expectedID, result.ID)
	assert.True(t, result.IsActive)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
