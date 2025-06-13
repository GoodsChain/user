package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewPostgresUserRepository tests the constructor
func TestNewPostgresUserRepository(t *testing.T) {
	db, _, repo := setupMockDB(t)
	defer db.Close()

	assert.NotNil(t, repo)
	assert.IsType(t, &postgresUserRepository{}, repo)
}
