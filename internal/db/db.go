package db

import (
	"fmt"
	"log"

	"github.com/GoodsChain/user/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// InitDB initializes and returns a database connection
func InitDB(cfg *config.Config) (*sqlx.DB, error) {
	connStr := cfg.GetDBConnectionString()
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ping the database to verify connection
	err = db.Ping()
	if err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to the database!")
	return db, nil
}
