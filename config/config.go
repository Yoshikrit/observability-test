package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Log      LogConfig
}

type AppConfig struct {
	Env  string
	Port string
}

type DatabaseConfig struct {
	DatabaseUrl string
}

type LogConfig struct {
	Level            string
	AccessLogEnabled bool
}

func Load() (*Config, error) {
	// .env is optional: in Docker/CI env vars are usually injected directly.
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			DatabaseUrl: getEnv("DATABASE_URL", ""),
		},
		Log: LogConfig{
			Level:            getEnv("LOG_LEVEL", "info"),
			AccessLogEnabled: getEnvBool("ACCESS_LOG_ENABLED", false),
		},
	}

	if cfg.Database.DatabaseUrl == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}
