package handler

import (
	"bytes"
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

// TestCreateUser_Success tests successful user creation with all fields
func TestCreateUser_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	createReq := models.CreateUserRequest{
		Email:    "test@example.com",
		FullName: "John Doe",
		Phone:    stringPtr("1234567890"),
		Role:     "admin",
	}

	expectedUser := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		FullName: "John Doe",
		Phone:    stringPtr("1234567890"),
		Role:     "admin",
		IsActive: true,
	}

	mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Email == "test@example.com" && user.FullName == "John Doe" && user.Role == "admin"
	})).Return(expectedUser, nil)

	requestBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, response.ID)
	assert.Equal(t, expectedUser.Email, response.Email)
	assert.Equal(t, expectedUser.FullName, response.FullName)
	assert.Equal(t, expectedUser.Phone, response.Phone)
	assert.Equal(t, expectedUser.Role, response.Role)

	mockRepo.AssertExpectations(t)
}

// TestCreateUser_SuccessMinimalFields tests user creation with minimal required fields
func TestCreateUser_SuccessMinimalFields(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	createReq := models.CreateUserRequest{
		Email:    "minimal@example.com",
		FullName: "Jane Doe",
		Role:     "staff",
		// Phone is optional
	}

	expectedUser := &models.User{
		ID:       uuid.New(),
		Email:    "minimal@example.com",
		FullName: "Jane Doe",
		Phone:    nil,
		Role:     "staff",
		IsActive: true,
	}

	mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Email == "minimal@example.com" && user.FullName == "Jane Doe" && user.Phone == nil
	})).Return(expectedUser, nil)

	requestBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.Email, response.Email)
	assert.Equal(t, expectedUser.FullName, response.FullName)
	assert.Nil(t, response.Phone)

	mockRepo.AssertExpectations(t)
}

// TestCreateUser_MalformedJSON tests handling of malformed JSON
func TestCreateUser_MalformedJSON(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("POST", "/api/v1/users/", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid character")
}

// TestCreateUser_InvalidEmail tests validation of invalid email format
func TestCreateUser_InvalidEmail(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	createReq := models.CreateUserRequest{
		Email:    "invalid-email", // Invalid email format
		FullName: "Test User",
		Role:     "admin",
	}

	requestBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "email")
}

// TestCreateUser_EmptyEmail tests validation of empty email
func TestCreateUser_EmptyEmail(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	createReq := models.CreateUserRequest{
		Email:    "", // Empty email
		FullName: "Test User",
		Role:     "admin",
	}

	requestBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "required")
}

// TestCreateUser_EmptyFullName tests validation of empty full name
func TestCreateUser_EmptyFullName(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	createReq := models.CreateUserRequest{
		Email:    "test@example.com",
		FullName: "", // Empty full name
		Role:     "admin",
	}

	requestBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "required")
}

// TestCreateUser_InvalidRole tests validation of invalid role
func TestCreateUser_InvalidRole(t *testing.T) {
	handler, _ := setupTestHandler()
	router := setupTestRouter(handler)

	createReq := models.CreateUserRequest{
		Email:    "test@example.com",
		FullName: "Test User",
		Role:     "invalid-role", // Invalid role
	}

	requestBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Role")
}

// TestCreateUser_RepositoryError tests handling of repository errors
func TestCreateUser_RepositoryError(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	createReq := models.CreateUserRequest{
		Email:    "error@example.com",
		FullName: "Error User",
		Role:     "admin",
	}

	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("database connection failed"))

	requestBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "database connection failed", response["error"])

	mockRepo.AssertExpectations(t)
}

// TestCreateUser_AllRoles tests creation with different valid roles
func TestCreateUser_AllRoles(t *testing.T) {
	validRoles := []string{"admin", "staff", "supplier"}

	for _, role := range validRoles {
		t.Run("Role_"+role, func(t *testing.T) {
			handler, mockRepo := setupTestHandler()
			router := setupTestRouter(handler)

			createReq := models.CreateUserRequest{
				Email:    role + "@example.com",
				FullName: role + " User",
				Role:     role,
			}

			expectedUser := &models.User{
				ID:       uuid.New(),
				Email:    role + "@example.com",
				FullName: role + " User",
				Role:     role,
				IsActive: true,
			}

			mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(expectedUser, nil)

			requestBody, _ := json.Marshal(createReq)
			req, _ := http.NewRequest("POST", "/api/v1/users/", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response models.User
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, role, response.Role)

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestCreateUser_DuplicateEmail tests handling of duplicate email errors
func TestCreateUser_DuplicateEmail(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	router := setupTestRouter(handler)

	createReq := models.CreateUserRequest{
		Email:    "duplicate@example.com",
		FullName: "Duplicate User",
		Role:     "admin",
	}

	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("duplicate key value violates unique constraint"))

	requestBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/users/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "duplicate key value violates unique constraint", response["error"])

	mockRepo.AssertExpectations(t)
}
