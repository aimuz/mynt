package share

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.aimuz.me/mynt/store"
)

func TestBStr(t *testing.T) {
	tests := []struct {
		input bool
		want  string
	}{
		{true, "yes"},
		{false, "no"},
	}
	for _, tt := range tests {
		if got := bStr(tt.input); got != tt.want {
			t.Errorf("bStr(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateShareSection(t *testing.T) {
	mgr := &Manager{}

	tests := []struct {
		name  string
		share store.Share
		want  string
	}{
		{
			name: "normal share with all options",
			share: store.Share{
				Name:       "projects",
				Path:       "/tank/projects",
				Comment:    "Project Files",
				ShareType:  store.ShareTypeNormal,
				ReadOnly:   false,
				Browseable: true,
				GuestOK:    false,
				ValidUsers: "alice,bob,charlie",
			},
			want: `[projects]
  path = /tank/projects
  comment = Project Files
  read only = no
  browseable = yes
  guest ok = no
  valid users = alice,bob,charlie
  create mask = 0664
  directory mask = 0775

`,
		},
		{
			name: "normal share read-only hidden with guest",
			share: store.Share{
				Name:       "test-normal",
				Path:       "/tank/normal",
				Comment:    "",
				ShareType:  store.ShareTypeNormal,
				ReadOnly:   true,
				Browseable: false,
				GuestOK:    true,
				ValidUsers: "user1,user2",
			},
			want: `[test-normal]
  path = /tank/normal
  comment = 
  read only = yes
  browseable = no
  guest ok = yes
  valid users = user1,user2
  create mask = 0664
  directory mask = 0775

`,
		},
		{
			name: "public share read-only",
			share: store.Share{
				Name:      "media",
				Path:      "/tank/media",
				Comment:   "Public Media Library",
				ShareType: store.ShareTypePublic,
				ReadOnly:  true,
			},
			want: `[media]
  path = /tank/media
  comment = Public Media Library
  browseable = yes
  guest ok = yes
  read only = yes
  create mask = 0644
  directory mask = 0755

`,
		},
		{
			name: "public share writable",
			share: store.Share{
				Name:      "upload",
				Path:      "/tank/upload",
				Comment:   "Public Upload Area",
				ShareType: store.ShareTypePublic,
				ReadOnly:  false,
			},
			want: `[upload]
  path = /tank/upload
  comment = Public Upload Area
  browseable = yes
  guest ok = yes
  read only = no
  create mask = 0644
  directory mask = 0755

`,
		},
		{
			name: "restricted share with valid users",
			share: store.Share{
				Name:       "finance",
				Path:       "/tank/finance",
				Comment:    "Finance Department Only",
				ShareType:  store.ShareTypeRestricted,
				ReadOnly:   false,
				ValidUsers: "admin,accountant,cfo",
			},
			want: `[finance]
  path = /tank/finance
  comment = Finance Department Only
  browseable = yes
  guest ok = no
  read only = no
  valid users = admin,accountant,cfo
  create mask = 0664
  directory mask = 0775

`,
		},
		{
			name: "restricted share without valid users",
			share: store.Share{
				Name:      "restricted-no-users",
				Path:      "/tank/restricted",
				Comment:   "Restricted without users",
				ShareType: store.ShareTypeRestricted,
				ReadOnly:  false,
			},
			want: `[restricted-no-users]
  path = /tank/restricted
  comment = Restricted without users
  browseable = yes
  guest ok = no
  read only = no
  create mask = 0664
  directory mask = 0775

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			mgr.generateShareSection(&buf, tt.share)
			got := buf.String()

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("generateShareSection() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGenerateSMBConfig(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	repo := store.NewShareRepo(db)

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
	}

	for i := range shares {
		if err := repo.Save(&shares[i]); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "smb.conf")
	mgr := NewManager(repo, configPath)

	if err := mgr.generateSMBConfig(); err != nil {
		t.Fatalf("generateSMBConfig: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	want := `[global]
  workgroup = WORKGROUP
  server string = Mynt NAS
  security = user
  map to guest = Bad User
  log file = /var/log/samba/%m.log
  max log size = 50

[finance]
  path = /tank/finance
  comment = Finance Dept
  browseable = yes
  guest ok = no
  read only = no
  valid users = admin,accountant
  create mask = 0664
  directory mask = 0775

[media]
  path = /tank/media
  comment = Public Media
  browseable = yes
  guest ok = yes
  read only = yes
  create mask = 0644
  directory mask = 0755

`

	if diff := cmp.Diff(want, string(data)); diff != "" {
		t.Errorf("generateSMBConfig() mismatch (-want +got):\n%s", diff)
	}
}

func TestNewManager_DefaultPaths(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	repo := store.NewShareRepo(db)

	tests := []struct {
		name       string
		configPath string
		wantEmpty  bool
	}{
		{"explicit path", "/custom/path.conf", false},
		{"empty uses default", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(repo, tt.configPath)
			if tt.configPath != "" && m.configPath != tt.configPath {
				t.Errorf("configPath = %q, want %q", m.configPath, tt.configPath)
			}
			if tt.configPath == "" && m.configPath == "" {
				t.Error("configPath should not be empty when not specified")
			}
		})
	}
}
