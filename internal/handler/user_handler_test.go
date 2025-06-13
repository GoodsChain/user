package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoodsChain/user/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, id uuid.UUID, updates *models.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context, filters *models.FilterParams, sort *models.SortParams, pagination *models.PaginationParams) (*models.GetUsersResponse, error) {
	args := m.Called(ctx, filters, sort, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GetUsersResponse), args.Error(1)
}

// setupTestHandler creates a test handler with mock repository
func setupTestHandler() (*UserHandler, *MockUserRepository) {
	mockRepo := &MockUserRepository{}
	handler := NewUserHandler(mockRepo)
	return handler, mockRepo
}

// setupTestRouter creates a test router with the handler
func setupTestRouter(handler *UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")
	users := v1.Group("/users")
	{
		users.GET("/", handler.GetAllUsers)
		users.POST("/", handler.CreateUser)
		users.GET("/:id", handler.GetUserByID)
		users.PATCH("/:id", handler.UpdateUser)
		users.DELETE("/:id", handler.DeleteUser)
	}
	return r
}

// TestUpdateUser_Success tests successful user update
func TestUpdateUser_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	updateReq := models.UpdateUserRequest{
		Email:    stringPtr("updated@example.com"),
		FullName: stringPtr("Updated Name"),
	}

	expectedUser := &models.User{
		ID:       userID,
		Email:    "updated@example.com",
		FullName: "Updated Name",
		Role:     "admin",
		IsActive: true,
	}

	mockRepo.On("UpdateUser", mock.Anything, userID, &updateReq).Return(expectedUser, nil)

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, response.ID)
	assert.Equal(t, expectedUser.Email, response.Email)
	assert.Equal(t, expectedUser.FullName, response.FullName)

	mockRepo.AssertExpectations(t)
}

// TestUpdateUser_InvalidUUID tests handling of invalid UUID
func TestUpdateUser_InvalidUUID(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	updateReq := models.UpdateUserRequest{
		Email: stringPtr("test@example.com"),
	}

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", "/api/v1/users/invalid-uuid", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid user ID format", response["error"])
}

// TestUpdateUser_MalformedJSON tests handling of malformed JSON
func TestUpdateUser_MalformedJSON(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid character")
}

// TestUpdateUser_NoFieldsProvided tests error when no fields are provided
func TestUpdateUser_NoFieldsProvided(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	updateReq := models.UpdateUserRequest{
		// No fields provided
	}

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "at least one field must be provided for update", response["error"])
}

// TestUpdateUser_ValidationError tests handling of validation errors
func TestUpdateUser_ValidationError(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	updateReq := models.UpdateUserRequest{
		Email: stringPtr("invalid-email"), // Invalid email format
	}

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "email") // Should contain validation error about email
}

// TestUpdateUser_UserNotFound tests handling when user doesn't exist
func TestUpdateUser_UserNotFound(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	updateReq := models.UpdateUserRequest{
		Email: stringPtr("test@example.com"),
	}

	mockRepo.On("UpdateUser", mock.Anything, userID, &updateReq).Return(nil, fmt.Errorf("user not found"))

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "user not found", response["error"])

	mockRepo.AssertExpectations(t)
}

// TestUpdateUser_DuplicateEmail tests handling of duplicate email errors
func TestUpdateUser_DuplicateEmail(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	updateReq := models.UpdateUserRequest{
		Email: stringPtr("duplicate@example.com"),
	}

	mockRepo.On("UpdateUser", mock.Anything, userID, &updateReq).Return(nil, fmt.Errorf("duplicate key value violates unique constraint"))

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "email already exists", response["error"])

	mockRepo.AssertExpectations(t)
}

// TestUpdateUser_DatabaseError tests handling of general database errors
func TestUpdateUser_DatabaseError(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	updateReq := models.UpdateUserRequest{
		Email: stringPtr("test@example.com"),
	}

	mockRepo.On("UpdateUser", mock.Anything, userID, &updateReq).Return(nil, fmt.Errorf("database connection failed"))

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "failed to update user", response["error"])

	mockRepo.AssertExpectations(t)
}

// TestUpdateUser_AllFields tests updating all fields
func TestUpdateUser_AllFields(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	updateReq := models.UpdateUserRequest{
		Email:    stringPtr("allfields@example.com"),
		FullName: stringPtr("All Fields User"),
		Phone:    stringPtr("1234567890"),
		Role:     stringPtr("supplier"),
		IsActive: boolPtr(false),
	}

	expectedUser := &models.User{
		ID:       userID,
		Email:    "allfields@example.com",
		FullName: "All Fields User",
		Phone:    stringPtr("1234567890"),
		Role:     "supplier",
		IsActive: false,
	}

	mockRepo.On("UpdateUser", mock.Anything, userID, &updateReq).Return(expectedUser, nil)

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, response.ID)
	assert.Equal(t, expectedUser.Email, response.Email)
	assert.Equal(t, expectedUser.FullName, response.FullName)
	assert.Equal(t, expectedUser.Phone, response.Phone)
	assert.Equal(t, expectedUser.Role, response.Role)
	assert.Equal(t, expectedUser.IsActive, response.IsActive)

	mockRepo.AssertExpectations(t)
}

// TestUpdateUser_InvalidRole tests validation of invalid role
func TestUpdateUser_InvalidRole(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	updateReq := models.UpdateUserRequest{
		Role: stringPtr("invalid-role"), // Invalid role
	}

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Role") // Should contain validation error about Role
}

// TestUpdateUser_EmptyFullName tests validation of empty full name
func TestUpdateUser_EmptyFullName(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	updateReq := models.UpdateUserRequest{
		FullName: stringPtr(""), // Empty full name
	}

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "min") // Should contain validation error about minimum length
}

// TestUpdateUser_NullPhone tests updating phone to null
func TestUpdateUser_NullPhone(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	updateReq := models.UpdateUserRequest{
		Phone: stringPtr(""), // Setting phone to empty (will be treated as null)
	}

	expectedUser := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		FullName: "Test User",
		Phone:    nil, // Phone set to null
		Role:     "admin",
		IsActive: true,
	}

	mockRepo.On("UpdateUser", mock.Anything, userID, &updateReq).Return(expectedUser, nil)

	requestBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%s", userID.String()), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, response.ID)
	assert.Nil(t, response.Phone)

	mockRepo.AssertExpectations(t)
}

// TestDeleteUser_Success tests successful user deletion
func TestDeleteUser_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	expectedUser := &models.User{
		ID:       userID,
		Email:    "delete@example.com",
		FullName: "Delete User",
		Role:     "admin",
		IsActive: true,
	}

	mockRepo.On("DeleteUser", mock.Anything, userID).Return(expectedUser, nil)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%s", userID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, response.ID)
	assert.Equal(t, expectedUser.Email, response.Email)
	assert.Equal(t, expectedUser.FullName, response.FullName)

	mockRepo.AssertExpectations(t)
}

// TestDeleteUser_InvalidUUID tests handling of invalid UUID
func TestDeleteUser_InvalidUUID(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("DELETE", "/api/v1/users/invalid-uuid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid user ID format", response["error"])
}

// TestDeleteUser_UserNotFound tests handling when user doesn't exist
func TestDeleteUser_UserNotFound(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	mockRepo.On("DeleteUser", mock.Anything, userID).Return(nil, fmt.Errorf("user not found"))

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%s", userID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "user not found", response["error"])

	mockRepo.AssertExpectations(t)
}

// TestDeleteUser_DatabaseError tests handling of general database errors
func TestDeleteUser_DatabaseError(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	mockRepo.On("DeleteUser", mock.Anything, userID).Return(nil, fmt.Errorf("database connection failed"))

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%s", userID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "failed to delete user", response["error"])

	mockRepo.AssertExpectations(t)
}

// TestGetUserByID_Success tests successful user retrieval
func TestGetUserByID_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	expectedUser := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		FullName: "Test User",
		Role:     "admin",
		IsActive: true,
	}

	mockRepo.On("GetUserByID", mock.Anything, userID).Return(expectedUser, nil)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", userID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, response.ID)
	assert.Equal(t, expectedUser.Email, response.Email)
	assert.Equal(t, expectedUser.FullName, response.FullName)

	mockRepo.AssertExpectations(t)
}

// TestGetUserByID_InvalidUUID tests handling of invalid UUID
func TestGetUserByID_InvalidUUID(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/v1/users/invalid-uuid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid user ID format", response["error"])
}

// TestGetUserByID_UserNotFound tests handling when user doesn't exist
func TestGetUserByID_UserNotFound(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	mockRepo.On("GetUserByID", mock.Anything, userID).Return(nil, fmt.Errorf("user not found"))

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", userID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "user not found", response["error"])

	mockRepo.AssertExpectations(t)
}

// TestGetUserByID_DatabaseError tests handling of general database errors
func TestGetUserByID_DatabaseError(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	userID := uuid.New()
	mockRepo.On("GetUserByID", mock.Anything, userID).Return(nil, fmt.Errorf("database connection failed"))

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", userID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "failed to retrieve user", response["error"])

	mockRepo.AssertExpectations(t)
}

// GetAllUsers Tests

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

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
