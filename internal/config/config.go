package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL      string
	ListenAddr       string
	JWTSecret        string
	CORSOrigins      []string
	WebhookAllowlist []string
	CSRFCookieName   string
	MaxTokenTTLSec   int64
}

func Load() (Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET must be set")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if strings.Contains(strings.ToLower(dbURL), "sslmode=disable") {
		return Config{}, fmt.Errorf("DATABASE_URL must use TLS (sslmode=require or verify-full)")
	}
	return Config{
		DatabaseURL:      dbURL,
		ListenAddr:       getenv("LISTEN_ADDR", "127.0.0.1:8081"),
		JWTSecret:        secret,
		CORSOrigins:      splitCSV(getenv("CORS_ORIGINS", "")),
		WebhookAllowlist: splitCSV(getenv("WEBHOOK_ALLOWLIST", "")),
		CSRFCookieName:   getenv("CSRF_COOKIE_NAME", "csrf_token"),
		MaxTokenTTLSec:   900,
	}, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
