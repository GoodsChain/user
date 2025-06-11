package router

import (
	"github.com/GoodsChain/user/internal/handler"
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up all the API routes
func SetupRouter(userHandler *handler.UserHandler) *gin.Engine {
	r := gin.Default()

	// API group for /api/v1
	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/", userHandler.CreateUser)
		}
	}

	return r
}
