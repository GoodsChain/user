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
