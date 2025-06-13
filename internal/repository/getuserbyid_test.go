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

// TestGetUserByID_Success tests successful user retrieval
func TestGetUserByID_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	expectedTime := time.Now()
	phone := "1234567890"

	// Expect get user query
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(userID, "get@example.com", "Get User", &phone, "admin", true, expectedTime, expectedTime))

	ctx := context.Background()
	result, err := repo.GetUserByID(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "get@example.com", result.Email)
	assert.Equal(t, "Get User", result.FullName)
	assert.Equal(t, &phone, result.Phone)
	assert.Equal(t, "admin", result.Role)
	assert.True(t, result.IsActive)
	assert.Equal(t, expectedTime, result.CreatedAt)
	assert.Equal(t, expectedTime, result.UpdatedAt)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetUserByID_UserNotFound tests error when user doesn't exist
func TestGetUserByID_UserNotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()

	// Expect get user query to fail
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	result, err := repo.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetUserByID_DatabaseError tests handling of database errors
func TestGetUserByID_DatabaseError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()

	// Expect get user query to fail with database error
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	result, err := repo.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetUserByID_ScanError tests handling of row scanning errors
func TestGetUserByID_ScanError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()

	// Return invalid data that will cause scanning to fail
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("invalid-uuid", "scan@example.com", "Scan User", nil, "admin", true, "invalid-time", "invalid-time"))

	ctx := context.Background()
	result, err := repo.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetUserByID_ContextCancelled tests handling of context cancellation
func TestGetUserByID_ContextCancelled(t *testing.T) {
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

	result, err := repo.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	// The error could be context.Canceled or a wrapped version
	assert.True(t, err == context.Canceled || err.Error() == "user not found: context canceled")
}

// TestGetUserByID_AllUserTypes tests retrieving users with different roles and data
func TestGetUserByID_AllUserTypes(t *testing.T) {
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

			// Expect get user query
			columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
			mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE id = \$1`).
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows(columns).
					AddRow(userID, email, fullName, phone, tc.role, tc.isActive, expectedTime, expectedTime))

			ctx := context.Background()
			result, err := repo.GetUserByID(ctx, userID)

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
