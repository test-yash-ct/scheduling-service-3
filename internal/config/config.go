package config

import (
	"os"

	"github.com/healthops/scheduling-service/internal/obs"
)

type Config struct {
	DatabaseURL string
	ListenAddr  string
	Metadata    obs.Metadata
}

func Load() Config {
	return Config{
		DatabaseURL: getenv("DATABASE_URL", "postgres://app:app@localhost:5432/scheduling?sslmode=disable"),
		ListenAddr:  getenv("LISTEN_ADDR", "0.0.0.0:8081"),
		Metadata:    obs.MetadataFromEnv("scheduling-service"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
