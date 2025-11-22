package store

import (
	"database/sql"
	"time"
)

// AccountType represents the type of user account.
type AccountType string

const (
	AccountVirtual AccountType = "virtual" // SMB/NFS only, no system access
	AccountSystem  AccountType = "system"  // Full Linux user with shell
)

// User represents a user account.
type User struct {
	ID           int64       `json:"id"`
	Username     string      `json:"username"`
	PasswordHash string      `json:"-"` // Never expose in JSON
	FullName     string      `json:"full_name"`
	Email        string      `json:"email"`
	AccountType  AccountType `json:"account_type"`
	IsAdmin      bool        `json:"is_admin"`
	IsActive     bool        `json:"is_active"`
	HomeDir      string      `json:"home_dir,omitempty"`
	Shell        string      `json:"shell,omitempty"`
	UID          *int        `json:"uid,omitempty"`
	GID          *int        `json:"gid,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	LastLogin    *time.Time  `json:"last_login,omitempty"`
}

// UserRepo manages user persistence.
type UserRepo struct {
	db *DB
}

// NewUserRepo creates a new user repository.
func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

// Save creates a new user.
func (r *UserRepo) Save(user *User) error {
	user.CreatedAt = time.Now()

	result, err := r.db.conn.Exec(`
		INSERT INTO users (username, password_hash, full_name, email, account_type, 
			is_admin, is_active, home_dir, shell, uid, gid, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user.Username, user.PasswordHash, user.FullName, user.Email, user.AccountType,
		user.IsAdmin, user.IsActive, user.HomeDir, user.Shell, user.UID, user.GID, user.CreatedAt)

	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	user.ID = id
	return nil
}

// List returns all users.
func (r *UserRepo) List() ([]User, error) {
	rows, err := r.db.conn.Query(`
		SELECT id, username, password_hash, full_name, email, account_type,
			is_admin, is_active, home_dir, shell, uid, gid, created_at, last_login
		FROM users
		ORDER BY username
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Email,
			&u.AccountType, &u.IsAdmin, &u.IsActive, &u.HomeDir, &u.Shell,
			&u.UID, &u.GID, &u.CreatedAt, &u.LastLogin)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

// GetByUsername retrieves a user by username.
func (r *UserRepo) GetByUsername(username string) (*User, error) {
	var u User
	err := r.db.conn.QueryRow(`
		SELECT id, username, password_hash, full_name, email, account_type,
			is_admin, is_active, home_dir, shell, uid, gid, created_at, last_login
		FROM users WHERE username = ?
	`, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Email,
		&u.AccountType, &u.IsAdmin, &u.IsActive, &u.HomeDir, &u.Shell,
		&u.UID, &u.GID, &u.CreatedAt, &u.LastLogin)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

// Update updates a user.
func (r *UserRepo) Update(user *User) error {
	_, err := r.db.conn.Exec(`
		UPDATE users SET full_name = ?, email = ?, is_admin = ?, is_active = ?
		WHERE id = ?
	`, user.FullName, user.Email, user.IsAdmin, user.IsActive, user.ID)
	return err
}

// UpdatePassword updates a user's password hash.
func (r *UserRepo) UpdatePassword(id int64, passwordHash string) error {
	_, err := r.db.conn.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, id)
	return err
}

// UpdateLastLogin updates the last login time.
func (r *UserRepo) UpdateLastLogin(id int64) error {
	now := time.Now()
	_, err := r.db.conn.Exec(`UPDATE users SET last_login = ? WHERE id = ?`, now, id)
	return err
}

// Delete removes a user.
func (r *UserRepo) Delete(id int64) error {
	_, err := r.db.conn.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}
