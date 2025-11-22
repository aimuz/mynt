package store

import (
	"time"

	"go.aimuz.me/mynt/disk"
)

// DiskState represents a persisted disk state.
type DiskState struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Model      string    `json:"model"`
	Serial     string    `json:"serial"`
	Size       uint64    `json:"size"`
	Type       string    `json:"type"`
	LastSeen   time.Time `json:"last_seen"`
	FirstSeen  time.Time `json:"first_seen"`
	IsAttached bool      `json:"is_attached"`
}

// DiskRepo manages disk state persistence.
type DiskRepo struct {
	db *DB
}

// NewDiskRepo creates a new disk repository.
func NewDiskRepo(db *DB) *DiskRepo {
	return &DiskRepo{db: db}
}

// Save or update a disk state.
func (r *DiskRepo) Save(info disk.Info) error {
	now := time.Now()

	// Check if disk exists
	var exists bool
	err := r.db.conn.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM disks WHERE name = ? AND serial = ?)",
		info.Name, info.Serial,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Update existing disk
		_, err = r.db.conn.Exec(`
			UPDATE disks 
			SET path = ?, model = ?, size = ?, type = ?, last_seen = ?, is_attached = 1
			WHERE name = ? AND serial = ?
		`, info.Path, info.Model, info.Size, string(info.Type), now, info.Name, info.Serial)
		return err
	}

	// Insert new disk
	_, err = r.db.conn.Exec(`
		INSERT INTO disks (name, path, model, serial, size, type, first_seen, last_seen, is_attached)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)
	`, info.Name, info.Path, info.Model, info.Serial, info.Size, string(info.Type), now, now)
	return err
}

// MarkDetached marks a disk as no longer attached.
func (r *DiskRepo) MarkDetached(name, serial string) error {
	_, err := r.db.conn.Exec(`
		UPDATE disks SET is_attached = 0, last_seen = ? WHERE name = ? AND serial = ?
	`, time.Now(), name, serial)
	return err
}

// ListAttached returns all currently attached disks.
func (r *DiskRepo) ListAttached() ([]DiskState, error) {
	rows, err := r.db.conn.Query(`
		SELECT name, path, model, serial, size, type, first_seen, last_seen, is_attached
		FROM disks
		WHERE is_attached = 1
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disks []DiskState
	for rows.Next() {
		var d DiskState
		if err := rows.Scan(
			&d.Name, &d.Path, &d.Model, &d.Serial, &d.Size, &d.Type,
			&d.FirstSeen, &d.LastSeen, &d.IsAttached,
		); err != nil {
			return nil, err
		}
		disks = append(disks, d)
	}
	return disks, nil
}

// GetBySerial retrieves a disk by its serial number.
func (r *DiskRepo) GetBySerial(serial string) (*DiskState, error) {
	var d DiskState
	err := r.db.conn.QueryRow(`
		SELECT name, path, model, serial, size, type, first_seen, last_seen, is_attached
		FROM disks
		WHERE serial = ?
	`, serial).Scan(
		&d.Name, &d.Path, &d.Model, &d.Serial, &d.Size, &d.Type,
		&d.FirstSeen, &d.LastSeen, &d.IsAttached,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ToInfo converts DiskState to disk.Info.
func (d *DiskState) ToInfo() disk.Info {
	return disk.Info{
		Name:   d.Name,
		Path:   d.Path,
		Model:  d.Model,
		Serial: d.Serial,
		Size:   d.Size,
		Type:   disk.Type(d.Type),
	}
}
