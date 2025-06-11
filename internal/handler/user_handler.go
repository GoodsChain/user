package handler

import (
	"net/http"
	"strings"

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
