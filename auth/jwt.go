// Package auth provides authentication and authorization utilities.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.aimuz.me/mynt/store"
)

// Claims represents JWT claims.
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// Config holds authentication configuration.
type Config struct {
	Secret         []byte
	TokenDuration  time.Duration
	RefreshEnabled bool
}

// DefaultConfig returns default authentication config.
func DefaultConfig(secret string) *Config {
	return &Config{
		Secret:         []byte(secret),
		TokenDuration:  24 * time.Hour,
		RefreshEnabled: false,
	}
}

// GenerateToken generates a JWT token for a user.
func GenerateToken(user *store.User, config *Config) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(config.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "mynt-nas",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.Secret)
}

// ValidateToken validates a JWT token and returns the claims.
func ValidateToken(tokenString string, config *Config) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return config.Secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
