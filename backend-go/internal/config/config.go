package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              int
	JWTSecret         string
	SessionTTLSeconds int
}

func Load() Config {
	return Config{
		Port:              envInt("PORT", 4000),
		JWTSecret:         envStr("JWT_SECRET", "dev-secret-change-me"),
		SessionTTLSeconds: envInt("SESSION_TTL_SECONDS", 60),
	}
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
