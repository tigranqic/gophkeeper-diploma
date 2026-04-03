// Package crypto provides AES-256-GCM encryption/decryption and scrypt-based
// key derivation for the GophKeeper client. All cryptographic operations use
// the standard library and golang.org/x/crypto.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/scrypt"
	"io"
)

// DeriveKeys derives DEK and AuthPass from master password and salt.
func DeriveKeys(password string, salt []byte) (dek []byte, authPass []byte, err error) {
	// 32 bytes for DEK, 32 bytes for AuthPass
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 64)
	if err != nil {
		return nil, nil, err
	}
	return key[:32], key[32:], nil
}

// Encrypt encrypts data using AES-GCM with dek.
func Encrypt(data []byte, dek []byte) ([]byte, error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-GCM with dek.
func Decrypt(ciphertext []byte, dek []byte) ([]byte, error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, encryptedData, nil)
}

// Zero clears the given byte slice by filling it with zeros.
func Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
