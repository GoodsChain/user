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
