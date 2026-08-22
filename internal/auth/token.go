package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Sub    string   `json:"sub"`
	Tenant string   `json:"tenant"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func ParseBearer(raw string, secret string, maxTTL time.Duration) (Claims, error) {
	if raw == "" {
		return Claims{}, errors.New("missing authorization")
	}
	parts := strings.SplitN(raw, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return Claims{}, errors.New("invalid authorization scheme")
	}
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || parsed == nil || !parsed.Valid {
		return Claims{}, errors.New("invalid token")
	}
	if claims.Sub == "" || claims.Tenant == "" {
		return Claims{}, errors.New("invalid subject")
	}
	if claims.ExpiresAt == nil || claims.IssuedAt == nil {
		return Claims{}, errors.New("token lifetime required")
	}
	if claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time) > maxTTL {
		return Claims{}, errors.New("token ttl exceeds policy")
	}
	return *claims, nil
}

func (c Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}
