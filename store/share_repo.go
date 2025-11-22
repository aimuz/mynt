package store

import (
	"database/sql"
	"time"
)

// ShareType represents the type of share access control
type ShareType string

const (
	ShareTypeNormal     ShareType = "normal"
	ShareTypePublic     ShareType = "public"
	ShareTypeRestricted ShareType = "restricted"
)

// Share represents a file share configuration.
type Share struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Protocol   string    `json:"protocol"` // smb, nfs
	ReadOnly   bool      `json:"read_only"`
	Browseable bool      `json:"browseable"`
	GuestOK    bool      `json:"guest_ok"`
	ValidUsers string    `json:"valid_users"` // comma-separated
	Comment    string    `json:"comment"`
	ShareType  ShareType `json:"share_type"` // normal, public, restricted
	CreatedAt  time.Time `json:"created_at"`
}

// ShareRepo manages share persistence.
type ShareRepo struct {
	db *DB
}

// NewShareRepo creates a new share repository.
func NewShareRepo(db *DB) *ShareRepo {
	return &ShareRepo{db: db}
}

// Save creates a new share.
func (r *ShareRepo) Save(share *Share) error {
	// Default to normal if not set
	if share.ShareType == "" {
		share.ShareType = ShareTypeNormal
	}

	share.CreatedAt = time.Now()

	result, err := r.db.conn.Exec(`
		INSERT INTO shares (name, path, protocol, read_only, browseable, guest_ok, valid_users, comment, share_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, share.Name, share.Path, share.Protocol, share.ReadOnly, share.Browseable,
		share.GuestOK, share.ValidUsers, share.Comment, share.ShareType, share.CreatedAt)

	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	share.ID = id
	return nil
}

// List returns all shares, optionally filtered by protocol.
func (r *ShareRepo) List(protocol string) ([]Share, error) {
	query := "SELECT id, name, path, protocol, read_only, browseable, guest_ok, valid_users, comment, share_type, created_at FROM shares"
	args := []any{}

	if protocol != "" {
		query += " WHERE protocol = ?"
		args = append(args, protocol)
	}

	query += " ORDER BY name"

	rows, err := r.db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []Share
	for rows.Next() {
		var s Share
		err := rows.Scan(&s.ID, &s.Name, &s.Path, &s.Protocol, &s.ReadOnly,
			&s.Browseable, &s.GuestOK, &s.ValidUsers, &s.Comment, &s.ShareType, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		shares = append(shares, s)
	}

	return shares, nil
}

// Get retrieves a share by ID.
func (r *ShareRepo) Get(id int64) (*Share, error) {
	var s Share
	err := r.db.conn.QueryRow(`
		SELECT id, name, path, protocol, read_only, browseable, guest_ok, valid_users, comment, share_type, created_at
		FROM shares WHERE id = ?
	`, id).Scan(&s.ID, &s.Name, &s.Path, &s.Protocol, &s.ReadOnly,
		&s.Browseable, &s.GuestOK, &s.ValidUsers, &s.Comment, &s.ShareType, &s.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

// Delete removes a share.
func (r *ShareRepo) Delete(id int64) error {
	_, err := r.db.conn.Exec("DELETE FROM shares WHERE id = ?", id)
	return err
}
