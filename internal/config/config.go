package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port       string
	StorageDir string
	MaxClients int

	LogLevel string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		StorageDir: getEnv("STORAGE_DIR", "./storage/"),
		MaxClients: 2,
		LogLevel:   "debug",
	}
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
