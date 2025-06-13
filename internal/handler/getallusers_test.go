package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoodsChain/user/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestGetAllUsers_Success tests successful retrieval with default parameters
func TestGetAllUsers_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	expectedUsers := []models.User{
		{
			ID:       uuid.New(),
			Email:    "user1@example.com",
			FullName: "User One",
			Role:     "admin",
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Email:    "user2@example.com",
			FullName: "User Two",
			Role:     "staff",
			IsActive: true,
		},
	}

	expectedResponse := &models.GetUsersResponse{
		Data: expectedUsers,
		Pagination: models.PaginationMetadata{
			Page:       1,
			PageSize:   10,
			Total:      2,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}

	mockRepo.On("GetAllUsers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/users/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, expectedResponse.Pagination.Total, response.Pagination.Total)
	assert.Equal(t, expectedResponse.Pagination.Page, response.Pagination.Page)

	mockRepo.AssertExpectations(t)
}

// TestGetAllUsers_WithFilters tests filtering functionality
func TestGetAllUsers_WithFilters(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	expectedUsers := []models.User{
		{
			ID:       uuid.New(),
			Email:    "admin@example.com",
			FullName: "Admin User",
			Role:     "admin",
			IsActive: true,
		},
	}

	expectedResponse := &models.GetUsersResponse{
		Data: expectedUsers,
		Pagination: models.PaginationMetadata{
			Page:       1,
			PageSize:   10,
			Total:      1,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}

	mockRepo.On("GetAllUsers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/users/?role=admin&is_active=true&search=admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, "admin", response.Data[0].Role)

	mockRepo.AssertExpectations(t)
}

// TestGetAllUsers_WithSorting tests sorting functionality
func TestGetAllUsers_WithSorting(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	expectedUsers := []models.User{
		{
			ID:       uuid.New(),
			Email:    "user2@example.com",
			FullName: "User Two",
			Role:     "staff",
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Email:    "user1@example.com",
			FullName: "User One",
			Role:     "admin",
			IsActive: true,
		},
	}

	expectedResponse := &models.GetUsersResponse{
		Data: expectedUsers,
		Pagination: models.PaginationMetadata{
			Page:       1,
			PageSize:   10,
			Total:      2,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}

	mockRepo.On("GetAllUsers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/users/?sort_by=email&sort_order=desc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 2)

	mockRepo.AssertExpectations(t)
}

// TestGetAllUsers_WithPagination tests pagination functionality
func TestGetAllUsers_WithPagination(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	expectedUsers := []models.User{
		{
			ID:       uuid.New(),
			Email:    "user3@example.com",
			FullName: "User Three",
			Role:     "staff",
			IsActive: true,
		},
	}

	expectedResponse := &models.GetUsersResponse{
		Data: expectedUsers,
		Pagination: models.PaginationMetadata{
			Page:       2,
			PageSize:   5,
			Total:      8,
			TotalPages: 2,
			HasNext:    false,
			HasPrev:    true,
		},
	}

	mockRepo.On("GetAllUsers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/users/?page=2&page_size=5", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 2, response.Pagination.Page)
	assert.Equal(t, 5, response.Pagination.PageSize)
	assert.Equal(t, 8, response.Pagination.Total)
	assert.True(t, response.Pagination.HasPrev)
	assert.False(t, response.Pagination.HasNext)

	mockRepo.AssertExpectations(t)
}

// TestGetAllUsers_WithDateFilters tests date filtering functionality
func TestGetAllUsers_WithDateFilters(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	expectedUsers := []models.User{
		{
			ID:       uuid.New(),
			Email:    "recent@example.com",
			FullName: "Recent User",
			Role:     "staff",
			IsActive: true,
		},
	}

	expectedResponse := &models.GetUsersResponse{
		Data: expectedUsers,
		Pagination: models.PaginationMetadata{
			Page:       1,
			PageSize:   10,
			Total:      1,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}

	mockRepo.On("GetAllUsers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/users/?created_from=2024-01-01&created_to=2024-12-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 1)

	mockRepo.AssertExpectations(t)
}

// TestGetAllUsers_InvalidDateFormat tests handling of invalid date formats
func TestGetAllUsers_InvalidDateFormat(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/v1/users/?created_from=invalid-date", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid created_from date format")
}

// TestGetAllUsers_InvalidSortField tests handling of invalid sort fields
func TestGetAllUsers_InvalidSortField(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/v1/users/?sort_by=invalid_field", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "SortBy")
}

// TestGetAllUsers_InvalidRole tests handling of invalid role values
func TestGetAllUsers_InvalidRole(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/v1/users/?role=invalid_role", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Role")
}

// TestGetAllUsers_InvalidPagination tests handling of invalid pagination parameters
func TestGetAllUsers_InvalidPagination(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/v1/users/?page=0&page_size=101", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Page")
}

// TestGetAllUsers_DatabaseError tests handling of database errors
func TestGetAllUsers_DatabaseError(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	mockRepo.On("GetAllUsers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("database connection failed"))

	req, _ := http.NewRequest("GET", "/api/v1/users/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "failed to retrieve users", response["error"])

	mockRepo.AssertExpectations(t)
}

// TestGetAllUsers_EmptyResult tests handling of empty results
func TestGetAllUsers_EmptyResult(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	expectedResponse := &models.GetUsersResponse{
		Data: []models.User{},
		Pagination: models.PaginationMetadata{
			Page:       1,
			PageSize:   10,
			Total:      0,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}

	mockRepo.On("GetAllUsers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/users/?search=nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 0)
	assert.Equal(t, 0, response.Pagination.Total)

	mockRepo.AssertExpectations(t)
}

// TestGetAllUsers_ComplexQuery tests a complex query with multiple filters, sorting, and pagination
func TestGetAllUsers_ComplexQuery(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	expectedUsers := []models.User{
		{
			ID:       uuid.New(),
			Email:    "active.admin@company.com",
			FullName: "Active Admin",
			Role:     "admin",
			IsActive: true,
		},
	}

	expectedResponse := &models.GetUsersResponse{
		Data: expectedUsers,
		Pagination: models.PaginationMetadata{
			Page:       1,
			PageSize:   5,
			Total:      1,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}

	mockRepo.On("GetAllUsers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/users/?role=admin&is_active=true&search=active&email_domain=company.com&sort_by=full_name&sort_order=asc&page=1&page_size=5", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, "admin", response.Data[0].Role)
	assert.True(t, response.Data[0].IsActive)

	mockRepo.AssertExpectations(t)
}
