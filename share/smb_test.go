package share

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.aimuz.me/mynt/store"
)

func TestGenerateShareSection_Normal(t *testing.T) {
	mgr := &Manager{}

	share := store.Share{
		Name:       "projects",
		Path:       "/tank/projects",
		Comment:    "Project Files",
		ShareType:  store.ShareTypeNormal,
		ReadOnly:   false,
		Browseable: true,
		GuestOK:    false,
		ValidUsers: "alice,bob,charlie",
	}

	var buf bytes.Buffer
	mgr.generateShareSection(&buf, share)

	config := buf.String()

	// Verify share name
	assert.Contains(t, config, "[projects]")

	// Verify path
	assert.Contains(t, config, "path = /tank/projects")

	// Verify comment
	assert.Contains(t, config, "comment = Project Files")

	// Verify normal share settings
	assert.Contains(t, config, "read only = no")
	assert.Contains(t, config, "browseable = yes")
	assert.Contains(t, config, "guest ok = no")
	assert.Contains(t, config, "valid users = alice,bob,charlie")
	assert.Contains(t, config, "create mask = 0664")
	assert.Contains(t, config, "directory mask = 0775")
}

func TestGenerateShareSection_Public(t *testing.T) {
	mgr := &Manager{}

	share := store.Share{
		Name:      "media",
		Path:      "/tank/media",
		Comment:   "Public Media Library",
		ShareType: store.ShareTypePublic,
		ReadOnly:  true,
	}

	var buf bytes.Buffer
	mgr.generateShareSection(&buf, share)

	config := buf.String()

	// Verify share name
	assert.Contains(t, config, "[media]")

	// Verify path
	assert.Contains(t, config, "path = /tank/media")

	// Verify public share settings
	assert.Contains(t, config, "browseable = yes")
	assert.Contains(t, config, "guest ok = yes")
	assert.Contains(t, config, "read only = yes")

	// Public shares should have specific permissions
	assert.Contains(t, config, "create mask = 0644")
	assert.Contains(t, config, "directory mask = 0755")

	// Public shares should not have valid users restriction
	assert.NotContains(t, config, "valid users")
}

func TestGenerateShareSection_Restricted(t *testing.T) {
	mgr := &Manager{}

	share := store.Share{
		Name:       "finance",
		Path:       "/tank/finance",
		Comment:    "Finance Department Only",
		ShareType:  store.ShareTypeRestricted,
		ReadOnly:   false,
		ValidUsers: "admin,accountant,cfo",
	}

	var buf bytes.Buffer
	mgr.generateShareSection(&buf, share)

	config := buf.String()

	// Verify share name
	assert.Contains(t, config, "[finance]")

	// Verify restricted share settings
	assert.Contains(t, config, "browseable = yes")
	assert.Contains(t, config, "guest ok = no")
	assert.Contains(t, config, "read only = no")
	assert.Contains(t, config, "valid users = admin,accountant,cfo")

	// Restricted shares should have group permissions
	assert.Contains(t, config, "create mask = 0664")
	assert.Contains(t, config, "directory mask = 0775")
}

func TestGenerateShareSection_RestrictedWithoutValidUsers(t *testing.T) {
	mgr := &Manager{}

	share := store.Share{
		Name:      "restricted-no-users",
		Path:      "/tank/restricted",
		Comment:   "Restricted without users",
		ShareType: store.ShareTypeRestricted,
		ReadOnly:  false,
	}

	var buf bytes.Buffer
	mgr.generateShareSection(&buf, share)

	config := buf.String()

	// Should still be restricted
	assert.Contains(t, config, "guest ok = no")

	// Should not have valid users line if empty
	assert.NotContains(t, config, "valid users =")
}

func TestGenerateShareSection_PublicReadWrite(t *testing.T) {
	mgr := &Manager{}

	share := store.Share{
		Name:      "upload",
		Path:      "/tank/upload",
		Comment:   "Public Upload Area",
		ShareType: store.ShareTypePublic,
		ReadOnly:  false, // Writable
	}

	var buf bytes.Buffer
	mgr.generateShareSection(&buf, share)

	config := buf.String()

	// Public but writable
	assert.Contains(t, config, "guest ok = yes")
	assert.Contains(t, config, "read only = no")
}

func TestBStr(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected string
	}{
		{"true to yes", true, "yes"},
		{"false to no", false, "no"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bStr(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSMBConfig_MultipleShares(t *testing.T) {
	// Create in-memory database for testing
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	repo := store.NewShareRepo(db)

	// Create test shares
	shares := []store.Share{
		{
			Name:      "media",
			Path:      "/tank/media",
			Protocol:  "smb",
			Comment:   "Public Media",
			ShareType: store.ShareTypePublic,
			ReadOnly:  true,
		},
		{
			Name:       "finance",
			Path:       "/tank/finance",
			Protocol:   "smb",
			Comment:    "Finance Dept",
			ShareType:  store.ShareTypeRestricted,
			ValidUsers: "admin,accountant",
		},
		{
			Name:       "projects",
			Path:       "/tank/projects",
			Protocol:   "smb",
			Comment:    "Project Files",
			ShareType:  store.ShareTypeNormal,
			Browseable: true,
			GuestOK:    false,
		},
	}

	for i := range shares {
		err := repo.Save(&shares[i])
		require.NoError(t, err)
	}

	// Create manager with temp config path
	mgr := NewManager(repo, "/tmp/test-smb.conf")

	// Generate config
	err = mgr.generateSMBConfig()
	require.NoError(t, err)

	// Read generated config
	// Note: In real scenario we would read the file, but for now we just verify no error
	// In future, we could mock the file system or use afero
}

func TestGenerateShareSection_AllShareTypes(t *testing.T) {
	mgr := &Manager{}

	testCases := []struct {
		name           string
		share          store.Share
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "normal_share_with_all_options",
			share: store.Share{
				Name:       "test-normal",
				Path:       "/tank/normal",
				ShareType:  store.ShareTypeNormal,
				ReadOnly:   true,
				Browseable: false,
				GuestOK:    true,
				ValidUsers: "user1,user2",
			},
			mustContain: []string{
				"[test-normal]",
				"read only = yes",
				"browseable = no",
				"guest ok = yes",
				"valid users = user1,user2",
			},
			mustNotContain: []string{},
		},
		{
			name: "public_share",
			share: store.Share{
				Name:      "test-public",
				Path:      "/tank/public",
				ShareType: store.ShareTypePublic,
			},
			mustContain: []string{
				"browseable = yes",
				"guest ok = yes",
			},
			mustNotContain: []string{
				"valid users",
			},
		},
		{
			name: "restricted_share",
			share: store.Share{
				Name:       "test-restricted",
				Path:       "/tank/restricted",
				ShareType:  store.ShareTypeRestricted,
				ValidUsers: "admin",
			},
			mustContain: []string{
				"browseable = yes",
				"guest ok = no",
				"valid users = admin",
			},
			mustNotContain: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			mgr.generateShareSection(&buf, tc.share)
			config := buf.String()

			for _, must := range tc.mustContain {
				assert.Contains(t, config, must, "Config should contain: %s", must)
			}

			for _, mustNot := range tc.mustNotContain {
				assert.NotContains(t, config, mustNot, "Config should not contain: %s", mustNot)
			}
		})
	}
}

func TestGenerateShareSection_ConfigFormat(t *testing.T) {
	mgr := &Manager{}

	share := store.Share{
		Name:      "format-test",
		Path:      "/tank/test",
		Comment:   "Format Test",
		ShareType: store.ShareTypeNormal,
	}

	var buf bytes.Buffer
	mgr.generateShareSection(&buf, share)

	config := buf.String()
	lines := strings.Split(config, "\n")

	// First line should be the section header
	assert.Equal(t, "[format-test]", lines[0])

	// Each config line should start with two spaces
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		// Skip empty lines
		if line == "" {
			continue
		}
		// Config lines should start with two spaces
		assert.True(t, strings.HasPrefix(line, "  "),
			"Line %d should start with two spaces: %q", i, line)
	}

	// Should end with a blank line
	assert.Equal(t, "", lines[len(lines)-1], "Last line should be empty")
}
