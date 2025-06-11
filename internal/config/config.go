package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configurations
type Config struct {
	Port string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// LoadConfig loads environment variables into the Config struct
func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file, using system environment variables: %v", err)
	}

	cfg := &Config{
		Port:       getEnv("PORT", "3000"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgre"),
		DBPassword: getEnv("DB_PASSWORD", "postgre"),
		DBName:     getEnv("DB_NAME", "user-database"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	return cfg, nil
}

// getEnv retrieves an environment variable, returning a default value if it's not set or empty.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDBConnectionString returns the PostgreSQL connection string
func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}
