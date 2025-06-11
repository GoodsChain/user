package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GoodsChain/user/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMockDB creates a mock database and sqlx wrapper for testing
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, UserRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPostgresUserRepository(sqlxDB)

	return db, mock, repo
}

// TestNewPostgresUserRepository tests the constructor
func TestNewPostgresUserRepository(t *testing.T) {
	db, _, repo := setupMockDB(t)
	defer db.Close()

	assert.NotNil(t, repo)
	assert.IsType(t, &postgresUserRepository{}, repo)
}

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

// TestDeleteUser_Success tests successful user deletion
func TestDeleteUser_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	expectedTime := time.Now()
	phone := "1234567890"

	// Expect user selection query
	selectColumns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(selectColumns).
			AddRow(userID, "delete@example.com", "Delete User", &phone, "admin", true, expectedTime, expectedTime))

	// Expect delete query
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	ctx := context.Background()
	result, err := repo.DeleteUser(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "delete@example.com", result.Email)
	assert.Equal(t, "Delete User", result.FullName)
	assert.Equal(t, &phone, result.Phone)
	assert.Equal(t, "admin", result.Role)
	assert.True(t, result.IsActive)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestDeleteUser_UserNotFound tests error when user doesn't exist for selection
func TestDeleteUser_UserNotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()

	// Expect user selection query to fail
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	result, err := repo.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestDeleteUser_DatabaseErrorOnSelect tests handling of database select errors
func TestDeleteUser_DatabaseErrorOnSelect(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()

	// Expect user selection query to fail with database error
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	result, err := repo.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestDeleteUser_DatabaseErrorOnDelete tests handling of database delete errors
func TestDeleteUser_DatabaseErrorOnDelete(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	expectedTime := time.Now()

	// Expect successful user selection
	selectColumns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(selectColumns).
			AddRow(userID, "delete@example.com", "Delete User", nil, "admin", true, expectedTime, expectedTime))

	// Expect delete query to fail
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	result, err := repo.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to delete user")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestDeleteUser_NoRowsAffected tests handling when delete affects no rows
func TestDeleteUser_NoRowsAffected(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	expectedTime := time.Now()

	// Expect successful user selection
	selectColumns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(selectColumns).
			AddRow(userID, "delete@example.com", "Delete User", nil, "admin", true, expectedTime, expectedTime))

	// Expect delete query to affect 0 rows (user was deleted between select and delete)
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	ctx := context.Background()
	result, err := repo.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestDeleteUser_RowsAffectedError tests handling when RowsAffected returns error
func TestDeleteUser_RowsAffectedError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	expectedTime := time.Now()

	// Expect successful user selection
	selectColumns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(selectColumns).
			AddRow(userID, "delete@example.com", "Delete User", nil, "admin", true, expectedTime, expectedTime))

	// Create a result that will error on RowsAffected
	result := sqlmock.NewErrorResult(sql.ErrConnDone)
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnResult(result)

	ctx := context.Background()
	deleteResult, err := repo.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, deleteResult)
	assert.Contains(t, err.Error(), "failed to get rows affected")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestDeleteUser_ScanError tests handling of row scanning errors
func TestDeleteUser_ScanError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()

	// Return invalid data that will cause scanning to fail
	selectColumns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(selectColumns).
			AddRow("invalid-uuid", "delete@example.com", "Delete User", nil, "admin", true, "invalid-time", "invalid-time"))

	ctx := context.Background()
	result, err := repo.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestDeleteUser_ContextCancelled tests handling of context cancellation
func TestDeleteUser_ContextCancelled(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The query should not be executed due to cancelled context
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(context.Canceled)

	result, err := repo.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	// The error could be context.Canceled or a wrapped version
	assert.True(t, err == context.Canceled || err.Error() == "user not found: context canceled")
}

// TestDeleteUser_AllUserTypes tests deleting users with different roles and data
func TestDeleteUser_AllUserTypes(t *testing.T) {
	testCases := []struct {
		name     string
		role     string
		hasPhone bool
		isActive bool
	}{
		{"AdminWithPhone", "admin", true, true},
		{"StaffWithoutPhone", "staff", false, true},
		{"SupplierInactive", "supplier", true, false},
		{"AdminInactive", "admin", false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, repo := setupMockDB(t)
			defer db.Close()

			userID := uuid.New()
			expectedTime := time.Now()
			var phone *string
			if tc.hasPhone {
				phoneStr := "1234567890"
				phone = &phoneStr
			}

			email := tc.role + "@example.com"
			fullName := tc.role + " User"

			// Expect user selection query
			selectColumns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
			mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows(selectColumns).
					AddRow(userID, email, fullName, phone, tc.role, tc.isActive, expectedTime, expectedTime))

			// Expect delete query
			mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
				WithArgs(userID).
				WillReturnResult(sqlmock.NewResult(0, 1))

			ctx := context.Background()
			result, err := repo.DeleteUser(ctx, userID)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, userID, result.ID)
			assert.Equal(t, email, result.Email)
			assert.Equal(t, fullName, result.FullName)
			assert.Equal(t, phone, result.Phone)
			assert.Equal(t, tc.role, result.Role)
			assert.Equal(t, tc.isActive, result.IsActive)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

// TestDeleteUser_ConcurrentDeletion tests the scenario where user is deleted between select and delete
func TestDeleteUser_ConcurrentDeletion(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	expectedTime := time.Now()

	// First query finds the user
	selectColumns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(selectColumns).
			AddRow(userID, "concurrent@example.com", "Concurrent User", nil, "admin", true, expectedTime, expectedTime))

	// But delete affects 0 rows (someone else deleted it)
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	ctx := context.Background()
	result, err := repo.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestUpdateUser_Success tests successful user update with single field
func TestUpdateUser_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	newEmail := "updated@example.com"
	updateReq := &models.UpdateUserRequest{
		Email: &newEmail,
	}

	expectedTime := time.Now()
	phone := "1234567890"

	// Expect user existence check
	mock.ExpectQuery(`SELECT id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

	// Expect update query
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`UPDATE users SET email = \$1, updated_at = \$2 WHERE id = \$3 RETURNING id, email, full_name, phone, role, is_active, created_at, updated_at`).
		WithArgs(newEmail, sqlmock.AnyArg(), userID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(userID, newEmail, "John Doe", &phone, "admin", true, expectedTime, expectedTime))

	ctx := context.Background()
	result, err := repo.UpdateUser(ctx, userID, updateReq)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, newEmail, result.Email)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestUpdateUser_SuccessMultipleFields tests successful user update with multiple fields
func TestUpdateUser_SuccessMultipleFields(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	newEmail := "multi@example.com"
	newFullName := "Updated Name"
	newPhone := "9876543210"
	newRole := "staff"
	isActive := false

	updateReq := &models.UpdateUserRequest{
		Email:    &newEmail,
		FullName: &newFullName,
		Phone:    &newPhone,
		Role:     &newRole,
		IsActive: &isActive,
	}

	expectedTime := time.Now()

	// Expect user existence check
	mock.ExpectQuery(`SELECT id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

	// Expect update query with all fields
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`UPDATE users SET email = \$1, full_name = \$2, phone = \$3, role = \$4, is_active = \$5, updated_at = \$6 WHERE id = \$7 RETURNING id, email, full_name, phone, role, is_active, created_at, updated_at`).
		WithArgs(newEmail, newFullName, newPhone, newRole, isActive, sqlmock.AnyArg(), userID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(userID, newEmail, newFullName, &newPhone, newRole, isActive, expectedTime, expectedTime))

	ctx := context.Background()
	result, err := repo.UpdateUser(ctx, userID, updateReq)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, newEmail, result.Email)
	assert.Equal(t, newFullName, result.FullName)
	assert.Equal(t, &newPhone, result.Phone)
	assert.Equal(t, newRole, result.Role)
	assert.Equal(t, isActive, result.IsActive)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestUpdateUser_UserNotFound tests error when user doesn't exist
func TestUpdateUser_UserNotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	newEmail := "notfound@example.com"
	updateReq := &models.UpdateUserRequest{
		Email: &newEmail,
	}

	// Expect user existence check to fail
	mock.ExpectQuery(`SELECT id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	result, err := repo.UpdateUser(ctx, userID, updateReq)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestUpdateUser_NoFieldsToUpdate tests error when no fields are provided
func TestUpdateUser_NoFieldsToUpdate(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	updateReq := &models.UpdateUserRequest{
		// No fields provided
	}

	// Expect user existence check
	mock.ExpectQuery(`SELECT id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

	ctx := context.Background()
	result, err := repo.UpdateUser(ctx, userID, updateReq)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no fields to update")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestUpdateUser_DatabaseError tests handling of database update errors
func TestUpdateUser_DatabaseError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	newEmail := "error@example.com"
	updateReq := &models.UpdateUserRequest{
		Email: &newEmail,
	}

	// Expect user existence check
	mock.ExpectQuery(`SELECT id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

	// Expect update query to fail
	mock.ExpectQuery(`UPDATE users SET email = \$1, updated_at = \$2 WHERE id = \$3 RETURNING`).
		WithArgs(newEmail, sqlmock.AnyArg(), userID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	result, err := repo.UpdateUser(ctx, userID, updateReq)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update user")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestUpdateUser_DuplicateEmail tests handling of email constraint violations
func TestUpdateUser_DuplicateEmail(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	duplicateEmail := "duplicate@example.com"
	updateReq := &models.UpdateUserRequest{
		Email: &duplicateEmail,
	}

	// Expect user existence check
	mock.ExpectQuery(`SELECT id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

	// Simulate unique constraint violation
	mock.ExpectQuery(`UPDATE users SET email = \$1, updated_at = \$2 WHERE id = \$3 RETURNING`).
		WithArgs(duplicateEmail, sqlmock.AnyArg(), userID).
		WillReturnError(sql.ErrNoRows) // Simulating constraint violation

	ctx := context.Background()
	result, err := repo.UpdateUser(ctx, userID, updateReq)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update user")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestUpdateUser_ScanError tests handling of row scanning errors
func TestUpdateUser_ScanError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	newEmail := "scan@example.com"
	updateReq := &models.UpdateUserRequest{
		Email: &newEmail,
	}

	// Expect user existence check
	mock.ExpectQuery(`SELECT id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

	// Return invalid data that will cause scanning to fail
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`UPDATE users SET email = \$1, updated_at = \$2 WHERE id = \$3 RETURNING`).
		WithArgs(newEmail, sqlmock.AnyArg(), userID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("invalid-uuid", newEmail, "Test User", nil, "admin", true, "invalid-time", "invalid-time"))

	ctx := context.Background()
	result, err := repo.UpdateUser(ctx, userID, updateReq)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to scan updated user")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestUpdateUser_NoRowsReturned tests handling when no rows are returned after update
func TestUpdateUser_NoRowsReturned(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	newEmail := "norows@example.com"
	updateReq := &models.UpdateUserRequest{
		Email: &newEmail,
	}

	// Expect user existence check
	mock.ExpectQuery(`SELECT id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

	// Return empty result set
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`UPDATE users SET email = \$1, updated_at = \$2 WHERE id = \$3 RETURNING`).
		WithArgs(newEmail, sqlmock.AnyArg(), userID).
		WillReturnRows(sqlmock.NewRows(columns)) // Empty rows

	ctx := context.Background()
	result, err := repo.UpdateUser(ctx, userID, updateReq)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no user returned after update")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestUpdateUser_AllFields tests updating all possible fields individually
func TestUpdateUser_AllFields(t *testing.T) {
	testCases := []struct {
		name   string
		req    *models.UpdateUserRequest
		column string
		value  interface{}
	}{
		{
			name:   "UpdateEmail",
			req:    &models.UpdateUserRequest{Email: stringPtr("newemail@example.com")},
			column: "email",
			value:  "newemail@example.com",
		},
		{
			name:   "UpdateFullName",
			req:    &models.UpdateUserRequest{FullName: stringPtr("New Name")},
			column: "full_name",
			value:  "New Name",
		},
		{
			name:   "UpdatePhone",
			req:    &models.UpdateUserRequest{Phone: stringPtr("5555555555")},
			column: "phone",
			value:  "5555555555",
		},
		{
			name:   "UpdateRole",
			req:    &models.UpdateUserRequest{Role: stringPtr("supplier")},
			column: "role",
			value:  "supplier",
		},
		{
			name:   "UpdateIsActive",
			req:    &models.UpdateUserRequest{IsActive: boolPtr(false)},
			column: "is_active",
			value:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, repo := setupMockDB(t)
			defer db.Close()

			userID := uuid.New()
			expectedTime := time.Now()
			phone := "1234567890"

			// Expect user existence check
			mock.ExpectQuery(`SELECT id FROM users WHERE id = \$1`).
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

			// Expect update query (pattern matching since exact query varies by field)
			columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
			mock.ExpectQuery(`UPDATE users SET .* WHERE id = .* RETURNING`).
				WithArgs(tc.value, sqlmock.AnyArg(), userID).
				WillReturnRows(sqlmock.NewRows(columns).
					AddRow(userID, "test@example.com", "Test User", &phone, "admin", true, expectedTime, expectedTime))

			ctx := context.Background()
			result, err := repo.UpdateUser(ctx, userID, tc.req)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, userID, result.ID)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
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
