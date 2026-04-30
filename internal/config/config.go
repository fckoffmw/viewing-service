package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type Config struct {
	Port                    string
	StorageDir              string
	MaxClients              int
	LogLevel                string
	LogFile                 string
	SessionsCleanupInterval int
}

func Load() *Config {
	return &Config{
		Port:                    getEnv("PORT", "8080"),
		StorageDir:              getEnv("STORAGE_DIR", "./storage/"),
		MaxClients:              getEnvInt("MAX_CLIENTS", 2),
		LogLevel:                getEnv("LOG_LEVEL", "debug"),
		LogFile:                 getEnv("LOG_FILE", ""),
		SessionsCleanupInterval: getEnvInt("SESSIONS_CLEANUP_INTERVAL", 300),
	}
}

func (c *Config) PrettyPrint() string {
	return fmt.Sprintf(
		"Port=%s StorageDir=%s MaxClients=%d LogLevel=%s LogFile=%s SessionsCleanupInterval=%d",
		c.Port, c.StorageDir, c.MaxClients, c.LogLevel, c.LogFile, c.SessionsCleanupInterval,
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
