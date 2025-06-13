package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/GoodsChain/user/internal/models"
	"github.com/GoodsChain/user/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// UserHandler handles HTTP requests related to users
type UserHandler struct {
	userRepo  repository.UserRepository
	validator *validator.Validate
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo:  userRepo,
		validator: validator.New(),
	}
}

// CreateUser handles the creation of a new user
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErrors.Error()})
		return
	}

	user := &models.User{
		Email:    req.Email,
		FullName: req.FullName,
		Phone:    req.Phone,
		Role:     req.Role,
	}

	createdUser, err := h.userRepo.CreateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}

// GetUserByID handles retrieving a user by their ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// Extract and validate user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Call repository to get user
	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		// Handle different error types
		errMsg := err.Error()
		if strings.Contains(errMsg, "user not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser handles updating an existing user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Extract and validate user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Bind JSON request body
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if at least one field is provided for update
	if req.Email == nil && req.FullName == nil && req.Phone == nil && req.Role == nil && req.IsActive == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided for update"})
		return
	}

	// Validate provided fields
	if err := h.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErrors.Error()})
		return
	}

	// Call repository to update user
	updatedUser, err := h.userRepo.UpdateUser(c.Request.Context(), userID, &req)
	if err != nil {
		// Handle different error types
		errMsg := err.Error()
		if strings.Contains(errMsg, "user not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if strings.Contains(errMsg, "no fields to update") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		if strings.Contains(errMsg, "duplicate key value") || strings.Contains(errMsg, "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// DeleteUser handles deleting an existing user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Extract and validate user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Call repository to delete user
	deletedUser, err := h.userRepo.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		// Handle different error types
		errMsg := err.Error()
		if strings.Contains(errMsg, "user not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, deletedUser)
}

// GetAllUsers handles retrieving all users with filtering, sorting, and pagination
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	// Parse query parameters
	var req models.GetUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate query parameters
	if err := h.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErrors.Error()})
		return
	}

	// Parse and convert parameters
	filters, sort, pagination, err := h.parseQueryParams(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call repository to get users
	response, err := h.userRepo.GetAllUsers(c.Request.Context(), filters, sort, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// parseQueryParams converts GetUsersRequest to repository parameters
func (h *UserHandler) parseQueryParams(req *models.GetUsersRequest) (*models.FilterParams, *models.SortParams, *models.PaginationParams, error) {
	// Build FilterParams
	filters := &models.FilterParams{
		Role:        req.Role,
		IsActive:    req.IsActive,
		Search:      req.Search,
		EmailDomain: req.EmailDomain,
	}

	// Parse date strings to time.Time
	if req.CreatedFrom != nil && *req.CreatedFrom != "" {
		createdFrom, err := time.Parse("2006-01-02", *req.CreatedFrom)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid created_from date format, use YYYY-MM-DD: %w", err)
		}
		filters.CreatedFrom = &createdFrom
	}

	if req.CreatedTo != nil && *req.CreatedTo != "" {
		createdTo, err := time.Parse("2006-01-02", *req.CreatedTo)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid created_to date format, use YYYY-MM-DD: %w", err)
		}
		// Set to end of day
		createdTo = createdTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filters.CreatedTo = &createdTo
	}

	if req.UpdatedFrom != nil && *req.UpdatedFrom != "" {
		updatedFrom, err := time.Parse("2006-01-02", *req.UpdatedFrom)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid updated_from date format, use YYYY-MM-DD: %w", err)
		}
		filters.UpdatedFrom = &updatedFrom
	}

	if req.UpdatedTo != nil && *req.UpdatedTo != "" {
		updatedTo, err := time.Parse("2006-01-02", *req.UpdatedTo)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid updated_to date format, use YYYY-MM-DD: %w", err)
		}
		// Set to end of day
		updatedTo = updatedTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filters.UpdatedTo = &updatedTo
	}

	// Build SortParams with defaults
	sortField := "created_at"
	sortOrder := "asc"
	
	if req.SortBy != nil {
		sortField = *req.SortBy
	}
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	sort := &models.SortParams{
		Field: sortField,
		Order: sortOrder,
	}

	// Build PaginationParams with defaults
	page := 1
	pageSize := 10

	if req.Page != nil {
		page = *req.Page
	}
	if req.PageSize != nil {
		pageSize = *req.PageSize
	}

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pagination := &models.PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}

	return filters, sort, pagination, nil
}
