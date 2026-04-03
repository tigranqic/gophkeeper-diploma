package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_Fields(t *testing.T) {
	id := uuid.New()
	u := User{
		ID:           id,
		Username:     "testuser",
		PasswordHash: "bcrypt-hash",
	}
	assert.Equal(t, id, u.ID)
	assert.Equal(t, "testuser", u.Username)
	assert.Equal(t, "bcrypt-hash", u.PasswordHash)
}
