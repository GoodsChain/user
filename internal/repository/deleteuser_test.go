package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
