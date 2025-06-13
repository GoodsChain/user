package handler

import (
	"context"

	"github.com/GoodsChain/user/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
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

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
