package repository

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// setupMockDB creates a mock database and sqlx wrapper for testing
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, UserRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPostgresUserRepository(sqlxDB)

	return db, mock, repo
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
