package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrypto_Roundtrip(t *testing.T) {
	password := "master-password"
	salt := []byte("constant-salt")
	data := []byte("sensitive-data")

	dek, _, err := DeriveKeys(password, salt)
	require.NoError(t, err)

	encrypted, err := Encrypt(data, dek)
	require.NoError(t, err)
	assert.NotEqual(t, data, encrypted)

	decrypted, err := Decrypt(encrypted, dek)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestCrypto_InvalidPassword(t *testing.T) {
	password := "master-password"
	salt := []byte("constant-salt")
	data := []byte("sensitive-data")

	dek, _, err := DeriveKeys(password, salt)
	require.NoError(t, err)

	encrypted, err := Encrypt(data, dek)
	require.NoError(t, err)

	wrongDek, _, err := DeriveKeys("wrong-password", salt)
	require.NoError(t, err)

	_, err = Decrypt(encrypted, wrongDek)
	assert.Error(t, err)
}

func TestCrypto_Zero(t *testing.T) {
	b := []byte("secret-data")
	Zero(b)
	for _, v := range b {
		assert.Equal(t, byte(0), v)
	}
}

func TestCrypto_Decrypt_ShortCiphertext(t *testing.T) {
	password := "master-password"
	salt := []byte("constant-salt")

	dek, _, err := DeriveKeys(password, salt)
	require.NoError(t, err)

	// A ciphertext shorter than the GCM nonce size must return an error.
	_, err = Decrypt([]byte("short"), dek)
	assert.Error(t, err)
}

func TestCrypto_DeriveKeys_ProducesDistinctDEKAndAuthPass(t *testing.T) {
	dek, authPass, err := DeriveKeys("password", []byte("salt"))
	require.NoError(t, err)
	assert.Len(t, dek, 32)
	assert.Len(t, authPass, 32)
	assert.NotEqual(t, dek, authPass)
}
