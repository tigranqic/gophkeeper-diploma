// Package config provides loading and parsing of application configuration
// from environment variables, command-line flags, and default values.
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// DefaultServerAddr is the default HTTP server listen address.
const DefaultServerAddr = "localhost:8080"

// DefaultGRPCAddr is the default gRPC server listen address.
const DefaultGRPCAddr = "localhost:3200"

// Config holds all runtime configuration parameters for the server.
type Config struct {
	LogLevel     string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	LogFormat    string        `yaml:"log_format" env:"LOG_FORMAT" env-default:"text"`
	ServerAddr   string        `yaml:"address" env:"ADDRESS" env-default:"localhost:8080"`
	GRPCAddr     string        `yaml:"grpc_address" env:"GRPC_ADDRESS" env-default:"localhost:3200"`
	DatabaseDSN  string        `yaml:"database_dsn" env:"DATABASE_DSN" env-default:"postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable"`
	JWTSecret    string        `yaml:"jwt_secret" env:"JWT_SECRET" env-default:"change-me-in-production"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"10s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"10s"`
}

// Load reads configuration from an optional config file (via -c flag or CONFIG env var),
// environment variables, and command-line flags. Flags take the highest precedence.
func Load(args []string) (*Config, error) {
	var cfg Config

	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		for i, arg := range args {
			if (arg == "-c" || arg == "-config") && i+1 < len(args) {
				configPath = args[i+1]
				break
			}
		}
	}

	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			return nil, fmt.Errorf("config file error: %w", err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, err
		}
	}

	fs := flag.NewFlagSet("gophkeeper", flag.ContinueOnError)
	fAddr := fs.String("a", cfg.ServerAddr, "server address")
	fGRPC := fs.String("g", cfg.GRPCAddr, "grpc address")
	fDSN := fs.String("d", cfg.DatabaseDSN, "database DSN")
	fLogLevel := fs.String("l", cfg.LogLevel, "log level")
	fJWTSecret := fs.String("jwt-secret", cfg.JWTSecret, "JWT signing secret")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.ServerAddr = *fAddr
		case "g":
			cfg.GRPCAddr = *fGRPC
		case "d":
			cfg.DatabaseDSN = *fDSN
		case "l":
			cfg.LogLevel = *fLogLevel
		case "jwt-secret":
			cfg.JWTSecret = *fJWTSecret
		}
	})

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	cfg.ServerAddr = strings.TrimPrefix(strings.TrimPrefix(cfg.ServerAddr, "https://"), "http://")
	cfg.GRPCAddr = strings.TrimPrefix(strings.TrimPrefix(cfg.GRPCAddr, "https://"), "http://")

	return &cfg, nil
}
