package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	StorageDir              string
	LogLevel                string
	LogFile                 string
	SessionsCleanupInterval int
	MaxRoomsPerUser         int
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:                    getEnv("PORT", "8080"),
		StorageDir:              getEnv("STORAGE_DIR", "./storage/"),
		LogLevel:                getEnv("LOG_LEVEL", "debug"),
		LogFile:                 getEnv("LOG_FILE", ""),
		SessionsCleanupInterval: getEnvInt("SESSIONS_CLEANUP_INTERVAL", 300),
		MaxRoomsPerUser:         getEnvInt("MAX_ROOMS_PER_USER", 10),
	}
}

func (c *Config) PrettyPrint() string {
	return fmt.Sprintf(
		"Port=%s StorageDir=%s LogLevel=%s LogFile=%s SessionsCleanupInterval=%d MaxRoomsPerUser=%d",
		c.Port, c.StorageDir, c.LogLevel, c.LogFile, c.SessionsCleanupInterval, c.MaxRoomsPerUser,
	)
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return i
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	if key == "PORT" {
		if _, err := strconv.Atoi(value); err != nil {
			return fallback
		}
	}
	return value
}
