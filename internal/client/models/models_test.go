package models

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPayload_MarshalUnmarshal(t *testing.T) {
	login := LoginPayload{
		Login:    "user",
		Password: "pass",
		Metadata: "meta",
	}

	data, err := MarshalPayload(login)
	require.NoError(t, err)

	var login2 LoginPayload
	err = UnmarshalPayload(data, &login2)
	require.NoError(t, err)

	assert.Equal(t, login, login2)
}
