package config

import "os"

type Config struct {
	DatabaseURL string
	ListenAddr  string
}

func Load() Config {
	return Config{
		DatabaseURL: getenv("DATABASE_URL", "postgres://app:app@localhost:5432/scheduling?sslmode=disable"),
		ListenAddr:  getenv("LISTEN_ADDR", "0.0.0.0:8081"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
