package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBPath             string
	LogPath            string
	SessionTimeout     time.Duration
	MaxFailedAttempts  int
	LockoutDuration    time.Duration
	ArgonMemory        uint32
	ArgonIterations    uint32
	ArgonParallelism   uint8
}

func Load() *Config {
	return &Config{
		DBPath:            getEnvString("DB_PATH", "./data/warden.db"),
		LogPath:           getEnvString("LOG_PATH", "./logs/warden.log"),
		SessionTimeout:    getEnvDuration("SESSION_TIMEOUT", 30*time.Minute),
		MaxFailedAttempts: getEnvInt("MAX_FAILED_ATTEMPTS", 5),
		LockoutDuration:   getEnvDuration("LOCKOUT_DURATION", 15*time.Minute),
		ArgonMemory:       uint32(getEnvInt("ARGON_MEMORY", 65536)),
		ArgonIterations:   uint32(getEnvInt("ARGON_ITERATIONS", 3)),
		ArgonParallelism:  uint8(getEnvInt("ARGON_PARALLELISM", 2)),
	}
}

func getEnvString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}