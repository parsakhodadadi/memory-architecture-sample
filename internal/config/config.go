package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort       string
	PostgresURL    string
	RedisAddress   string
	RedisPassword  string
	ShortTermTTL   time.Duration
	ShortTermLimit int64
	LongTermTTL    time.Duration
}

func Load() Config {
	return Config{
		HTTPPort:       value("HTTP_PORT", "8080"),
		PostgresURL:    value("POSTGRES_URL", "postgres://memory:memory@localhost:5433/memory?sslmode=disable"),
		RedisAddress:   value("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		ShortTermTTL:   duration("SHORT_TERM_TTL", 30*time.Minute),
		ShortTermLimit: integer("SHORT_TERM_LIMIT", 20),
		LongTermTTL:    duration("LONG_TERM_TTL", 30*24*time.Hour),
	}
}

func value(key, fallback string) string {
	if result := os.Getenv(key); result != "" {
		return result
	}
	return fallback
}

func duration(key string, fallback time.Duration) time.Duration {
	result, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return result
}

func integer(key string, fallback int64) int64 {
	result, err := strconv.ParseInt(os.Getenv(key), 10, 64)
	if err != nil || result <= 0 {
		return fallback
	}
	return result
}
