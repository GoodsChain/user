package main

import (
	"log"
	"net/http"

	"github.com/GoodsChain/user/internal/config"
	"github.com/GoodsChain/user/internal/db"
	"github.com/GoodsChain/user/internal/handler"
	"github.com/GoodsChain/user/internal/repository"
	"github.com/GoodsChain/user/internal/router"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize database connection
	db, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	// Initialize repository and handler
	userRepo := repository.NewPostgresUserRepository(db)
	userHandler := handler.NewUserHandler(userRepo)

	// Setup router
	r := router.SetupRouter(userHandler)

	// Start the server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
