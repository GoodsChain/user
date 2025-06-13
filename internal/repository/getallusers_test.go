package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GoodsChain/user/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetAllUsers_Success tests successful retrieval with basic pagination
func TestGetAllUsers_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Setup test data
	user1ID := uuid.New()
	user2ID := uuid.New()
	expectedTime := time.Now()
	phone := "1234567890"

	// Expect count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Expect data query
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users ORDER BY created_at ASC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(user1ID, "user1@example.com", "User One", &phone, "admin", true, expectedTime, expectedTime).
			AddRow(user2ID, "user2@example.com", "User Two", nil, "staff", true, expectedTime, expectedTime))

	filters := &models.FilterParams{}
	sort := &models.SortParams{Field: "created_at", Order: "asc"}
	pagination := &models.PaginationParams{Page: 1, PageSize: 10, Offset: 0}

	ctx := context.Background()
	result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 2)
	
	// Check first user
	assert.Equal(t, user1ID, result.Data[0].ID)
	assert.Equal(t, "user1@example.com", result.Data[0].Email)
	assert.Equal(t, "User One", result.Data[0].FullName)
	assert.Equal(t, &phone, result.Data[0].Phone)
	assert.Equal(t, "admin", result.Data[0].Role)
	assert.True(t, result.Data[0].IsActive)

	// Check second user
	assert.Equal(t, user2ID, result.Data[1].ID)
	assert.Equal(t, "user2@example.com", result.Data[1].Email)
	assert.Equal(t, "User Two", result.Data[1].FullName)
	assert.Nil(t, result.Data[1].Phone)
	assert.Equal(t, "staff", result.Data[1].Role)

	// Check pagination metadata
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.PageSize)
	assert.Equal(t, 2, result.Pagination.Total)
	assert.Equal(t, 1, result.Pagination.TotalPages)
	assert.False(t, result.Pagination.HasNext)
	assert.False(t, result.Pagination.HasPrev)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetAllUsers_WithFilters tests retrieval with various filters
func TestGetAllUsers_WithFilters(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	expectedTime := time.Now()
	role := "admin"
	isActive := true
	search := "john"

	// Expect count query with filters
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE role = \$1 AND is_active = \$2 AND \(LOWER\(full_name\) LIKE LOWER\(\$3\) OR LOWER\(email\) LIKE LOWER\(\$3\)\)`).
		WithArgs(role, isActive, "%"+search+"%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Expect data query with filters
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE role = \$1 AND is_active = \$2 AND \(LOWER\(full_name\) LIKE LOWER\(\$3\) OR LOWER\(email\) LIKE LOWER\(\$3\)\) ORDER BY created_at ASC LIMIT \$4 OFFSET \$5`).
		WithArgs(role, isActive, "%"+search+"%", 10, 0).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(userID, "john@example.com", "John Doe", nil, "admin", true, expectedTime, expectedTime))

	filters := &models.FilterParams{
		Role:     &role,
		IsActive: &isActive,
		Search:   &search,
	}
	sort := &models.SortParams{Field: "created_at", Order: "asc"}
	pagination := &models.PaginationParams{Page: 1, PageSize: 10, Offset: 0}

	ctx := context.Background()
	result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, userID, result.Data[0].ID)
	assert.Equal(t, "john@example.com", result.Data[0].Email)
	assert.Equal(t, "admin", result.Data[0].Role)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetAllUsers_WithTimeFilters tests retrieval with time-based filters
func TestGetAllUsers_WithTimeFilters(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	expectedTime := time.Now()
	createdFrom := time.Now().AddDate(0, 0, -7) // 7 days ago
	createdTo := time.Now()

	// Expect count query with time filters
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE created_at >= \$1 AND created_at <= \$2`).
		WithArgs(createdFrom, createdTo).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Expect data query with time filters
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users WHERE created_at >= \$1 AND created_at <= \$2 ORDER BY created_at ASC LIMIT \$3 OFFSET \$4`).
		WithArgs(createdFrom, createdTo, 10, 0).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(userID, "recent@example.com", "Recent User", nil, "staff", true, expectedTime, expectedTime))

	filters := &models.FilterParams{
		CreatedFrom: &createdFrom,
		CreatedTo:   &createdTo,
	}
	sort := &models.SortParams{Field: "created_at", Order: "asc"}
	pagination := &models.PaginationParams{Page: 1, PageSize: 10, Offset: 0}

	ctx := context.Background()
	result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "recent@example.com", result.Data[0].Email)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetAllUsers_WithSorting tests different sorting options
func TestGetAllUsers_WithSorting(t *testing.T) {
	testCases := []struct {
		name      string
		sortField string
		sortOrder string
		expectedQuery string
	}{
		{"SortByEmailDesc", "email", "desc", "ORDER BY email DESC"},
		{"SortByNameAsc", "full_name", "asc", "ORDER BY full_name ASC"},
		{"SortByRoleDesc", "role", "desc", "ORDER BY role DESC"},
		{"SortByCreatedAtAsc", "created_at", "asc", "ORDER BY created_at ASC"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, repo := setupMockDB(t)
			defer db.Close()

			userID := uuid.New()
			expectedTime := time.Now()

			// Expect count query
			mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

			// Expect data query with specific sorting
			columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
			mock.ExpectQuery(fmt.Sprintf(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users %s LIMIT \$1 OFFSET \$2`, tc.expectedQuery)).
				WithArgs(10, 0).
				WillReturnRows(sqlmock.NewRows(columns).
					AddRow(userID, "test@example.com", "Test User", nil, "admin", true, expectedTime, expectedTime))

			filters := &models.FilterParams{}
			sort := &models.SortParams{Field: tc.sortField, Order: tc.sortOrder}
			pagination := &models.PaginationParams{Page: 1, PageSize: 10, Offset: 0}

			ctx := context.Background()
			result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result.Data, 1)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

// TestGetAllUsers_WithPagination tests pagination scenarios
func TestGetAllUsers_WithPagination(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Test second page with 5 items per page, total 12 items
	userID := uuid.New()
	expectedTime := time.Now()

	// Expect count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(12))

	// Expect data query for page 2 (offset 5, limit 5)
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users ORDER BY created_at ASC LIMIT \$1 OFFSET \$2`).
		WithArgs(5, 5).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(userID, "page2@example.com", "Page Two User", nil, "staff", true, expectedTime, expectedTime))

	filters := &models.FilterParams{}
	sort := &models.SortParams{Field: "created_at", Order: "asc"}
	pagination := &models.PaginationParams{Page: 2, PageSize: 5, Offset: 5}

	ctx := context.Background()
	result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 1)

	// Check pagination metadata
	assert.Equal(t, 2, result.Pagination.Page)
	assert.Equal(t, 5, result.Pagination.PageSize)
	assert.Equal(t, 12, result.Pagination.Total)
	assert.Equal(t, 3, result.Pagination.TotalPages) // ceil(12/5) = 3
	assert.True(t, result.Pagination.HasNext)       // Page 2 of 3, so has next
	assert.True(t, result.Pagination.HasPrev)       // Page 2 of 3, so has prev

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetAllUsers_EmptyResult tests handling empty result set
func TestGetAllUsers_EmptyResult(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Expect count query returning 0
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Expect data query returning empty set
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users ORDER BY created_at ASC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(columns))

	filters := &models.FilterParams{}
	sort := &models.SortParams{Field: "created_at", Order: "asc"}
	pagination := &models.PaginationParams{Page: 1, PageSize: 10, Offset: 0}

	ctx := context.Background()
	result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 0)
	assert.Equal(t, 0, result.Pagination.Total)
	assert.Equal(t, 1, result.Pagination.TotalPages) // Minimum 1 page
	assert.False(t, result.Pagination.HasNext)
	assert.False(t, result.Pagination.HasPrev)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetAllUsers_CountError tests error handling for count query
func TestGetAllUsers_CountError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Expect count query to fail
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnError(sql.ErrConnDone)

	filters := &models.FilterParams{}
	sort := &models.SortParams{Field: "created_at", Order: "asc"}
	pagination := &models.PaginationParams{Page: 1, PageSize: 10, Offset: 0}

	ctx := context.Background()
	result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get total count")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetAllUsers_DataQueryError tests error handling for data query
func TestGetAllUsers_DataQueryError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Expect successful count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Expect data query to fail
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users ORDER BY created_at ASC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnError(sql.ErrConnDone)

	filters := &models.FilterParams{}
	sort := &models.SortParams{Field: "created_at", Order: "asc"}
	pagination := &models.PaginationParams{Page: 1, PageSize: 10, Offset: 0}

	ctx := context.Background()
	result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get users")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetAllUsers_ScanError tests error handling for row scanning
func TestGetAllUsers_ScanError(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Expect successful count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Expect data query with invalid data that will cause scanning to fail
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users ORDER BY created_at ASC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("invalid-uuid", "scan@example.com", "Scan User", nil, "admin", true, "invalid-time", "invalid-time"))

	filters := &models.FilterParams{}
	sort := &models.SortParams{Field: "created_at", Order: "asc"}
	pagination := &models.PaginationParams{Page: 1, PageSize: 10, Offset: 0}

	ctx := context.Background()
	result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get users")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetAllUsers_NilParameters tests handling of nil parameters
func TestGetAllUsers_NilParameters(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.New()
	expectedTime := time.Now()

	// Expect count query without WHERE clause (nil filters)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Expect data query with default sorting (nil sort params)
	columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users ORDER BY created_at ASC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(userID, "nil@example.com", "Nil Test User", nil, "admin", true, expectedTime, expectedTime))

	ctx := context.Background()
	result, err := repo.GetAllUsers(ctx, nil, nil, &models.PaginationParams{Page: 1, PageSize: 10, Offset: 0})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "nil@example.com", result.Data[0].Email)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestGetAllUsers_ContextCancelled tests handling of context cancellation
func TestGetAllUsers_ContextCancelled(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Expect count query to fail due to cancelled context
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnError(context.Canceled)

	filters := &models.FilterParams{}
	sort := &models.SortParams{Field: "created_at", Order: "asc"}
	pagination := &models.PaginationParams{Page: 1, PageSize: 10, Offset: 0}

	result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	// The error could be context.Canceled or a wrapped version
	assert.True(t, err == context.Canceled || err.Error() == "failed to get total count: context canceled")
}

// TestGetAllUsers_EdgeCasePagination tests edge cases in pagination calculation
func TestGetAllUsers_EdgeCasePagination(t *testing.T) {
	testCases := []struct {
		name           string
		total          int
		pageSize       int
		page           int
		expectedPages  int
		expectedHasNext bool
		expectedHasPrev bool
	}{
		{"ExactDivision", 20, 10, 1, 2, true, false},
		{"LastPage", 15, 5, 3, 3, false, true},
		{"SinglePage", 5, 10, 1, 1, false, false},
		{"ZeroTotal", 0, 10, 1, 1, false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, repo := setupMockDB(t)
			defer db.Close()

			// Expect count query
			mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tc.total))

			// Expect data query (may return empty for zero total)
			columns := []string{"id", "email", "full_name", "phone", "role", "is_active", "created_at", "updated_at"}
			mock.ExpectQuery(`SELECT id, email, full_name, phone, role, is_active, created_at, updated_at FROM users ORDER BY created_at ASC LIMIT \$1 OFFSET \$2`).
				WithArgs(tc.pageSize, (tc.page-1)*tc.pageSize).
				WillReturnRows(sqlmock.NewRows(columns))

			filters := &models.FilterParams{}
			sort := &models.SortParams{Field: "created_at", Order: "asc"}
			pagination := &models.PaginationParams{
				Page:     tc.page,
				PageSize: tc.pageSize,
				Offset:   (tc.page - 1) * tc.pageSize,
			}

			ctx := context.Background()
			result, err := repo.GetAllUsers(ctx, filters, sort, pagination)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.total, result.Pagination.Total)
			assert.Equal(t, tc.expectedPages, result.Pagination.TotalPages)
			assert.Equal(t, tc.expectedHasNext, result.Pagination.HasNext)
			assert.Equal(t, tc.expectedHasPrev, result.Pagination.HasPrev)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
