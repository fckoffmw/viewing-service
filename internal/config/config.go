package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type Config struct {
	Port       string
	StorageDir string
	MaxClients int
	LogLevel   string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		StorageDir: getEnv("STORAGE_DIR", "./storage/"),
		MaxClients: getEnvInt("MAX_CLIENTS", 2),
		LogLevel:   getEnv("LOG_LEVEL", "debug"),
	}
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
