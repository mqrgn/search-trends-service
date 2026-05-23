package config

import (
	"os"
)

type Config struct {
	NatsURL     string
	NatsSubject string
	HTTPPort    string
}

func Load() *Config {
	return &Config{
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),
		NatsSubject: getEnv("NATS_SUBJECT", "search.events"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
