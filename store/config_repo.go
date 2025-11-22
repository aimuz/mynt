package store

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

// ConfigRepo manages system configuration.
type ConfigRepo struct {
	db *DB
}

// NewConfigRepo creates a new config repository.
func NewConfigRepo(db *DB) *ConfigRepo {
	return &ConfigRepo{db: db}
}

// Get retrieves a config value.
func (r *ConfigRepo) Get(key string) (string, error) {
	var value string
	err := r.db.conn.QueryRow(`SELECT value FROM system_config WHERE key = ?`, key).Scan(&value)
	return value, err
}

// Set saves a config value.
func (r *ConfigRepo) Set(key, value string) error {
	now := time.Now()
	_, err := r.db.conn.Exec(`
		INSERT INTO system_config (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`, key, value, now, value, now)
	return err
}

// IsInitialized checks if the system has been initialized.
func (r *ConfigRepo) IsInitialized() (bool, error) {
	_, err := r.Get("initialized")
	if err != nil {
		return false, nil
	}
	return true, nil
}

// MarkInitialized marks the system as initialized.
func (r *ConfigRepo) MarkInitialized() error {
	return r.Set("initialized", "true")
}

// GetJWTSecret retrieves or generates the JWT secret.
func (r *ConfigRepo) GetJWTSecret() (string, error) {
	secret, err := r.Get("jwt_secret")
	if err != nil {
		// Generate new secret
		secret, err = generateRandomSecret(32)
		if err != nil {
			return "", err
		}
		// Save it
		if err := r.Set("jwt_secret", secret); err != nil {
			return "", err
		}
	}
	return secret, nil
}

// generateRandomSecret generates a random base64 encoded secret.
func generateRandomSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
