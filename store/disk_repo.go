package store

import (
	"encoding/json"
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

// SmartState represents cached SMART data.
type SmartState struct {
	DiskName            string           `json:"disk_name"`
	Passed              bool             `json:"passed"`
	Temperature         int              `json:"temperature"`
	PowerOnHours        int64            `json:"power_on_hours"`
	PowerCycleCount     int64            `json:"power_cycle_count"`
	ReallocatedSectors  int64            `json:"reallocated_sectors"`
	PendingSectors      int64            `json:"pending_sectors"`
	UncorrectableErrors int64            `json:"uncorrectable_errors"`
	Attributes          []disk.Attribute `json:"attributes"`
	UpdatedAt           time.Time        `json:"updated_at"`
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

// SaveSmart saves or updates SMART data for a disk.
func (r *DiskRepo) SaveSmart(report *disk.DetailedReport) error {
	attrs, _ := json.Marshal(report.Attributes)

	_, err := r.db.conn.Exec(`
		INSERT INTO disk_smart (disk_name, passed, temperature, power_on_hours, power_cycle_count,
			reallocated_sectors, pending_sectors, uncorrectable_errors, attributes, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(disk_name) DO UPDATE SET
			passed = excluded.passed,
			temperature = excluded.temperature,
			power_on_hours = excluded.power_on_hours,
			power_cycle_count = excluded.power_cycle_count,
			reallocated_sectors = excluded.reallocated_sectors,
			pending_sectors = excluded.pending_sectors,
			uncorrectable_errors = excluded.uncorrectable_errors,
			attributes = excluded.attributes,
			updated_at = excluded.updated_at
	`, report.Disk, report.Passed, report.Temperature, report.PowerOnHours, report.PowerCycleCount,
		report.ReallocatedSectors, report.PendingSectors, report.UncorrectableErrors, attrs, time.Now())
	return err
}

// GetSmart retrieves cached SMART data for a disk.
func (r *DiskRepo) GetSmart(name string) (*SmartState, error) {
	var s SmartState
	var attrsJSON []byte

	err := r.db.conn.QueryRow(`
		SELECT disk_name, passed, temperature, power_on_hours, power_cycle_count,
			reallocated_sectors, pending_sectors, uncorrectable_errors, attributes, updated_at
		FROM disk_smart WHERE disk_name = ?
	`, name).Scan(
		&s.DiskName, &s.Passed, &s.Temperature, &s.PowerOnHours, &s.PowerCycleCount,
		&s.ReallocatedSectors, &s.PendingSectors, &s.UncorrectableErrors, &attrsJSON, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(attrsJSON) > 0 {
		json.Unmarshal(attrsJSON, &s.Attributes)
	}
	return &s, nil
}

// ListSmart returns all cached SMART data.
func (r *DiskRepo) ListSmart() (map[string]*SmartState, error) {
	rows, err := r.db.conn.Query(`
		SELECT disk_name, passed, temperature, power_on_hours, power_cycle_count,
			reallocated_sectors, pending_sectors, uncorrectable_errors, attributes, updated_at
		FROM disk_smart
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*SmartState)
	for rows.Next() {
		var s SmartState
		var attrsJSON []byte
		if err := rows.Scan(
			&s.DiskName, &s.Passed, &s.Temperature, &s.PowerOnHours, &s.PowerCycleCount,
			&s.ReallocatedSectors, &s.PendingSectors, &s.UncorrectableErrors, &attrsJSON, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if len(attrsJSON) > 0 {
			json.Unmarshal(attrsJSON, &s.Attributes)
		}
		result[s.DiskName] = &s
	}
	return result, nil
}

// DeleteSmart removes SMART data for a disk.
func (r *DiskRepo) DeleteSmart(name string) error {
	_, err := r.db.conn.Exec("DELETE FROM disk_smart WHERE disk_name = ?", name)
	return err
}

// SmartCacheAdapter adapts DiskRepo to disk.SmartCache interface.
type SmartCacheAdapter struct {
	repo *DiskRepo
}

// NewSmartCache creates a SmartCache from DiskRepo.
func (r *DiskRepo) NewSmartCache() *SmartCacheAdapter {
	return &SmartCacheAdapter{repo: r}
}

// GetSmart implements disk.SmartCache.
func (a *SmartCacheAdapter) GetSmart(name string) (*disk.CachedSmart, error) {
	s, err := a.repo.GetSmart(name)
	if err != nil {
		return nil, err
	}
	return &disk.CachedSmart{
		Passed:              s.Passed,
		Temperature:         s.Temperature,
		ReallocatedSectors:  s.ReallocatedSectors,
		PendingSectors:      s.PendingSectors,
		UncorrectableErrors: s.UncorrectableErrors,
	}, nil
}

// ListSmart implements disk.SmartCache.
func (a *SmartCacheAdapter) ListSmart() (map[string]*disk.CachedSmart, error) {
	list, err := a.repo.ListSmart()
	if err != nil {
		return nil, err
	}
	result := make(map[string]*disk.CachedSmart, len(list))
	for k, v := range list {
		result[k] = &disk.CachedSmart{
			Passed:              v.Passed,
			Temperature:         v.Temperature,
			ReallocatedSectors:  v.ReallocatedSectors,
			PendingSectors:      v.PendingSectors,
			UncorrectableErrors: v.UncorrectableErrors,
		}
	}
	return result, nil
}
