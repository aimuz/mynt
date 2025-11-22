package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShareRepo_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := NewShareRepo(db)

	share := &Share{
		Name:       "testshare",
		Path:       "/tank/test",
		Protocol:   "smb",
		ReadOnly:   false,
		Browseable: true,
		GuestOK:    false,
		ValidUsers: "user1,user2",
	}

	err := repo.Save(share)
	require.NoError(t, err)
	require.Greater(t, share.ID, int64(0))
}

func TestShareRepo_Get(t *testing.T) {
	db := setupTestDB(t)
	repo := NewShareRepo(db)

	// Create share
	share := &Share{
		Name:     "gettest",
		Path:     "/tank/gettest",
		Protocol: "smb",
	}
	repo.Save(share)

	// Get by ID
	retrieved, err := repo.Get(share.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, share.Name, retrieved.Name)
	require.Equal(t, share.Path, retrieved.Path)
}

func TestShareRepo_Get_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewShareRepo(db)

	share, err := repo.Get(999)
	require.NoError(t, err)
	require.Nil(t, share)
}

func TestShareRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewShareRepo(db)

	// Create multiple shares
	shares := []*Share{
		{Name: "share1", Path: "/tank/share1", Protocol: "smb"},
		{Name: "share2", Path: "/tank/share2", Protocol: "nfs"},
		{Name: "share3", Path: "/tank/share3", Protocol: "smb"},
	}

	for _, s := range shares {
		repo.Save(s)
	}

	// List SMB shares
	list, err := repo.List("smb")
	require.NoError(t, err)
	require.Len(t, list, 2)

	// List all shares
	allList, err := repo.List("")
	require.NoError(t, err)
	require.Len(t, allList, 3)
}

func TestShareRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewShareRepo(db)

	share := &Share{
		Name:     "deltest",
		Path:     "/tank/deltest",
		Protocol: "smb",
	}
	repo.Save(share)

	// Delete
	err := repo.Delete(share.ID)
	require.NoError(t, err)

	// Verify deleted
	retrieved, _ := repo.Get(share.ID)
	require.Nil(t, retrieved)
}
