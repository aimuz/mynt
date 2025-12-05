package store

import (
	"database/sql"
	"encoding/json"
	"time"
)

// SnapshotPolicy represents a snapshot schedule policy.
type SnapshotPolicy struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Schedule  string    `json:"schedule"`  // e.g., "@daily", "0 * * * *"
	Retention string    `json:"retention"` // e.g., "7d", "24h"
	Datasets  []string  `json:"datasets"`  // List of dataset names
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SnapshotPolicyRepo manages snapshot policy persistence.
type SnapshotPolicyRepo struct {
	db *DB
}

// NewSnapshotPolicyRepo creates a new snapshot policy repository.
func NewSnapshotPolicyRepo(db *DB) *SnapshotPolicyRepo {
	return &SnapshotPolicyRepo{db: db}
}

// Save creates a new snapshot policy.
func (r *SnapshotPolicyRepo) Save(policy *SnapshotPolicy) error {
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	datasetsJSON, err := json.Marshal(policy.Datasets)
	if err != nil {
		return err
	}

	result, err := r.db.conn.Exec(`
		INSERT INTO snapshot_policies (name, schedule, retention, datasets, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, policy.Name, policy.Schedule, policy.Retention, string(datasetsJSON), policy.Enabled, policy.CreatedAt, policy.UpdatedAt)

	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	policy.ID = id
	return nil
}

// Update updates an existing snapshot policy.
func (r *SnapshotPolicyRepo) Update(policy *SnapshotPolicy) error {
	policy.UpdatedAt = time.Now()

	datasetsJSON, err := json.Marshal(policy.Datasets)
	if err != nil {
		return err
	}

	_, err = r.db.conn.Exec(`
		UPDATE snapshot_policies 
		SET name = ?, schedule = ?, retention = ?, datasets = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`, policy.Name, policy.Schedule, policy.Retention, string(datasetsJSON), policy.Enabled, policy.UpdatedAt, policy.ID)

	return err
}

// List returns all snapshot policies.
func (r *SnapshotPolicyRepo) List() ([]SnapshotPolicy, error) {
	query := "SELECT id, name, schedule, retention, datasets, enabled, created_at, updated_at FROM snapshot_policies ORDER BY name"

	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []SnapshotPolicy
	for rows.Next() {
		var p SnapshotPolicy
		var datasetsJSON string
		err := rows.Scan(&p.ID, &p.Name, &p.Schedule, &p.Retention, &datasetsJSON, &p.Enabled, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if datasetsJSON != "" {
			_ = json.Unmarshal([]byte(datasetsJSON), &p.Datasets)
		}
		if p.Datasets == nil {
			p.Datasets = []string{}
		}

		policies = append(policies, p)
	}

	return policies, nil
}

// Get retrieves a snapshot policy by ID.
func (r *SnapshotPolicyRepo) Get(id int64) (*SnapshotPolicy, error) {
	var p SnapshotPolicy
	var datasetsJSON string

	err := r.db.conn.QueryRow(`
		SELECT id, name, schedule, retention, datasets, enabled, created_at, updated_at
		FROM snapshot_policies WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Schedule, &p.Retention, &datasetsJSON, &p.Enabled, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if datasetsJSON != "" {
		_ = json.Unmarshal([]byte(datasetsJSON), &p.Datasets)
	}
	if p.Datasets == nil {
		p.Datasets = []string{}
	}

	return &p, nil
}

// Delete removes a snapshot policy.
func (r *SnapshotPolicyRepo) Delete(id int64) error {
	_, err := r.db.conn.Exec("DELETE FROM snapshot_policies WHERE id = ?", id)
	return err
}
