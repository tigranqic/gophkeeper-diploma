package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadEnv(t *testing.T) {
	os.Setenv("ADDRESS", "localhost:9999")
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("ADDRESS")
	defer os.Unsetenv("LOG_LEVEL")

	cfg, err := Load([]string{})
	require.NoError(t, err)
	assert.Equal(t, "localhost:9999", cfg.ServerAddr)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestConfig_LoadFlags(t *testing.T) {
	cfg, err := Load([]string{"-a", "0.0.0.0:7777", "-g", "0.0.0.0:4400", "-l", "warn"})
	require.NoError(t, err)
	assert.Equal(t, "0.0.0.0:7777", cfg.ServerAddr)
	assert.Equal(t, "0.0.0.0:4400", cfg.GRPCAddr)
	assert.Equal(t, "warn", cfg.LogLevel)
}

func TestConfig_Defaults(t *testing.T) {
	// Ensure no env vars leak from other tests.
	os.Unsetenv("ADDRESS")
	os.Unsetenv("GRPC_ADDRESS")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("JWT_SECRET")

	cfg, err := Load([]string{})
	require.NoError(t, err)
	assert.Equal(t, DefaultServerAddr, cfg.ServerAddr)
	assert.Equal(t, DefaultGRPCAddr, cfg.GRPCAddr)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.NotEmpty(t, cfg.JWTSecret)
}

func TestConfig_URLStripping(t *testing.T) {
	cfg, err := Load([]string{"-a", "http://localhost:8080", "-g", "https://localhost:3200"})
	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", cfg.ServerAddr)
	assert.Equal(t, "localhost:3200", cfg.GRPCAddr)
}

func TestConfig_JWTSecretFlag(t *testing.T) {
	cfg, err := Load([]string{"-jwt-secret", "my-super-secret"})
	require.NoError(t, err)
	assert.Equal(t, "my-super-secret", cfg.JWTSecret)
}

func TestConfig_JWTSecretEnv(t *testing.T) {
	os.Setenv("JWT_SECRET", "env-secret")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load([]string{})
	require.NoError(t, err)
	assert.Equal(t, "env-secret", cfg.JWTSecret)
}
