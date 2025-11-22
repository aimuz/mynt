package store

import (
	"encoding/json"
	"time"

	"go.aimuz.me/mynt/event"
)

// NotificationStatus represents the status of a notification.
type NotificationStatus string

const (
	NotificationUnread NotificationStatus = "unread"
	NotificationRead   NotificationStatus = "read"
	NotificationAcked  NotificationStatus = "acknowledged"
)

// Notification represents a persisted event notification.
type Notification struct {
	ID        int64              `json:"id"`
	Type      string             `json:"type"`
	Data      string             `json:"data"` // JSON encoded
	Status    NotificationStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	ReadAt    *time.Time         `json:"read_at,omitempty"`
	AckedAt   *time.Time         `json:"acked_at,omitempty"`
}

// NotificationRepo manages event notification persistence.
type NotificationRepo struct {
	db *DB
}

// NewNotificationRepo creates a new notification repository.
func NewNotificationRepo(db *DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

// Save persists an event as a notification.
func (r *NotificationRepo) Save(evt event.Event) error {
	data, err := json.Marshal(evt.Data)
	if err != nil {
		return err
	}

	_, err = r.db.conn.Exec(`
		INSERT INTO notifications (type, data, status, created_at)
		VALUES (?, ?, ?, ?)
	`, evt.Type, string(data), NotificationUnread, evt.Time)
	return err
}

// List retrieves notifications with filters.
func (r *NotificationRepo) List(status NotificationStatus, limit, offset int) ([]Notification, error) {
	query := `
		SELECT id, type, data, status, created_at, read_at, acked_at
		FROM notifications
	`
	args := []any{}

	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(
			&n.ID, &n.Type, &n.Data, &n.Status,
			&n.CreatedAt, &n.ReadAt, &n.AckedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

// MarkRead marks a notification as read.
func (r *NotificationRepo) MarkRead(id int64) error {
	now := time.Now()
	_, err := r.db.conn.Exec(`
		UPDATE notifications 
		SET status = ?, read_at = ?
		WHERE id = ? AND status = ?
	`, NotificationRead, now, id, NotificationUnread)
	return err
}

// MarkAcknowledged marks a notification as acknowledged (processed).
func (r *NotificationRepo) MarkAcknowledged(id int64) error {
	now := time.Now()
	_, err := r.db.conn.Exec(`
		UPDATE notifications 
		SET status = ?, acked_at = ?
		WHERE id = ?
	`, NotificationAcked, now, id)
	return err
}

// Delete removes a notification.
func (r *NotificationRepo) Delete(id int64) error {
	_, err := r.db.conn.Exec(`DELETE FROM notifications WHERE id = ?`, id)
	return err
}

// Count returns the number of notifications by status.
func (r *NotificationRepo) Count(status NotificationStatus) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM notifications`
	args := []any{}

	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}

	err := r.db.conn.QueryRow(query, args...).Scan(&count)
	return count, err
}
