package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.aimuz.me/mynt/event"
)

func TestNotificationRepo_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepo(db)

	evt := event.Event{
		Type: "disk.detected",
		Time: time.Now(),
		Data: map[string]string{"disk": "sda"},
	}

	err := repo.Save(evt)
	require.NoError(t, err)
}

func TestNotificationRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepo(db)

	// Create multiple notifications
	for i := 0; i < 5; i++ {
		evt := event.Event{
			Type: "test.event",
			Time: time.Now(),
		}
		repo.Save(evt)
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	list, err := repo.List("", 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 5)

	// Should be in reverse chronological order
	require.True(t, list[0].CreatedAt.After(list[4].CreatedAt))
}

func TestNotificationRepo_List_WithStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepo(db)

	// Create notifications
	evt1 := event.Event{Type: "test1", Time: time.Now()}
	evt2 := event.Event{Type: "test2", Time: time.Now()}
	repo.Save(evt1)
	repo.Save(evt2)

	// All should be unread initially
	list, err := repo.List(NotificationUnread, 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 2)

	// List read should be empty
	readList, err := repo.List(NotificationRead, 10, 0)
	require.NoError(t, err)
	require.Len(t, readList, 0)
}

func TestNotificationRepo_List_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepo(db)

	// Create 10 notifications
	for i := 0; i < 10; i++ {
		evt := event.Event{Type: "test", Time: time.Now()}
		repo.Save(evt)
		time.Sleep(time.Millisecond)
	}

	// Get first page
	page1, _ := repo.List("", 5, 0)
	require.Len(t, page1, 5)

	// Get second page
	page2, _ := repo.List("", 5, 5)
	require.Len(t, page2, 5)

	// Should be different
	require.NotEqual(t, page1[0].ID, page2[0].ID)
}

func TestNotificationRepo_MarkRead(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepo(db)

	evt := event.Event{Type: "test", Time: time.Now()}
	repo.Save(evt)

	// Get the notification ID
	list, _ := repo.List("", 1, 0)
	require.Len(t, list, 1)
	notif := list[0]

	require.Equal(t, NotificationUnread, notif.Status)

	err := repo.MarkRead(notif.ID)
	require.NoError(t, err)

	// Verify it's marked as read
	readList, _ := repo.List(NotificationRead, 10, 0)
	require.Len(t, readList, 1)
	require.NotNil(t, readList[0].ReadAt)
}

func TestNotificationRepo_MarkAcknowledged(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepo(db)

	evt := event.Event{Type: "test", Time: time.Now()}
	repo.Save(evt)

	list, err := repo.List("", 1, 0)
	require.NoError(t, err)
	if len(list) == 0 {
		t.Fatal("list is empty")
	}
	notif := list[0]

	err = repo.MarkAcknowledged(notif.ID)
	require.NoError(t, err)

	// Verify
	ackedList, err := repo.List(NotificationAcked, 10, 0)
	require.Len(t, ackedList, 1)
	require.NotNil(t, ackedList[0].AckedAt)
}

func TestNotificationRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepo(db)

	evt := event.Event{Type: "test", Time: time.Now()}
	err := repo.Save(evt)
	require.NoError(t, err)

	list, err := repo.List("", 1, 0)
	require.NoError(t, err)
	if len(list) == 0 {
		t.Fatal("list is empty")
	}
	notif := list[0]

	err = repo.Delete(notif.ID)
	require.NoError(t, err)

	// Verify deleted
	afterDelete, _ := repo.List("", 10, 0)
	require.Len(t, afterDelete, 0)
}

func TestNotificationRepo_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNotificationRepo(db)

	// Create some notifications
	for i := 0; i < 3; i++ {
		evt := event.Event{Type: "test", Time: time.Now()}
		repo.Save(evt)
		time.Sleep(time.Millisecond)
	}

	// All unread
	count, err := repo.Count(NotificationUnread)
	require.NoError(t, err)
	require.Equal(t, 3, count)

	// Mark one as read
	list, _ := repo.List("", 1, 0)
	repo.MarkRead(list[0].ID)

	// Check counts
	unreadCount, _ := repo.Count(NotificationUnread)
	require.Equal(t, 2, unreadCount)

	readCount, _ := repo.Count(NotificationRead)
	require.Equal(t, 1, readCount)

	totalCount, _ := repo.Count("")
	require.Equal(t, 3, totalCount)
}
