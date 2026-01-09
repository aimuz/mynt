package disk

import (
	"testing"
)

func TestParseTemperature(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "simple",
			input: "35",
			want:  35,
		},
		{
			name:  "with_units",
			input: "42 (Min/Max 20/55)",
			want:  42,
		},
		{
			name:  "complex",
			input: "38 (0 15 0 0 0)",
			want:  38,
		},
		{
			name:  "empty",
			input: "",
			want:  0,
		},
		{
			name:  "invalid",
			input: "not_a_number",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTemperature(tt.input)
			if got != tt.want {
				t.Errorf("parseTemperature(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTestType_Constants(t *testing.T) {
	tests := []struct {
		typ  TestType
		want string
	}{
		{TestShort, "short"},
		{TestLong, "long"},
	}

	for _, tt := range tests {
		t.Run(string(tt.typ), func(t *testing.T) {
			if string(tt.typ) != tt.want {
				t.Errorf("TestType = %v, want %v", tt.typ, tt.want)
			}
		})
	}
}

func TestSmartExitCodes(t *testing.T) {
	tests := []struct {
		name       string
		exitCode   int
		wantFatal  bool
	}{
		{
			name:      "success",
			exitCode:  0,
			wantFatal: false,
		},
		{
			name:      "cmd_line_error",
			exitCode:  smartExitCmdLine,
			wantFatal: true,
		},
		{
			name:      "dev_open_error",
			exitCode:  smartExitDevOpen,
			wantFatal: true,
		},
		{
			name:      "cmd_failed",
			exitCode:  smartExitCmdFailed,
			wantFatal: true,
		},
		{
			name:      "disk_failing",
			exitCode:  1 << 3, // Bit 3: DISK FAILING
			wantFatal: false,
		},
		{
			name:      "prefail_attributes",
			exitCode:  1 << 4, // Bit 4: Prefail attributes
			wantFatal: false,
		},
		{
			name:      "error_log",
			exitCode:  1 << 6, // Bit 6: Error log
			wantFatal: false,
		},
		{
			name:      "combined_non_fatal",
			exitCode:  (1 << 3) | (1 << 4) | (1 << 5),
			wantFatal: false,
		},
		{
			name:      "combined_with_fatal",
			exitCode:  (1 << 3) | smartExitDevOpen,
			wantFatal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isFatal := (tt.exitCode & smartExitFatalMask) != 0
			if isFatal != tt.wantFatal {
				t.Errorf("exit code %d: isFatal = %v, want %v", tt.exitCode, isFatal, tt.wantFatal)
			}
		})
	}
}

func TestAttributeIDs(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{"ReallocatedSectors", attrReallocatedSectors},
		{"PowerOnHours", attrPowerOnHours},
		{"Temperature", attrTemperature},
		{"PendingSectors", attrPendingSectors},
		{"Uncorrectable", attrUncorrectable},
	}

	// Verify expected values
	expected := map[string]int{
		"ReallocatedSectors": 5,
		"PowerOnHours":       9,
		"Temperature":        194,
		"PendingSectors":     197,
		"Uncorrectable":      198,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := expected[tt.name]
			if tt.id != want {
				t.Errorf("attribute %s: id = %v, want %v", tt.name, tt.id, want)
			}
		})
	}
}
