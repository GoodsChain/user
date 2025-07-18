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
