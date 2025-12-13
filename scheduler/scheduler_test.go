package scheduler

import (
	"testing"
)

func TestConvertSchedule(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "hourly",
			input: "@hourly",
			want:  "0 0 * * * *",
		},
		{
			name:  "daily",
			input: "@daily",
			want:  "0 0 0 * * *",
		},
		{
			name:  "weekly",
			input: "@weekly",
			want:  "0 0 0 * * 0",
		},
		{
			name:  "monthly",
			input: "@monthly",
			want:  "0 0 0 1 * *",
		},
		{
			name:  "5_field_cron",
			input: "0 0 * * *",
			want:  "0 0 0 * * *",
		},
		{
			name:  "6_field_cron",
			input: "0 0 0 * * *",
			want:  "0 0 0 * * *",
		},
		{
			name:  "custom_5_field",
			input: "30 2 * * 1",
			want:  "0 30 2 * * 1",
		},
		{
			name:  "already_6_field",
			input: "0 30 2 * * 1",
			want:  "0 30 2 * * 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertSchedule(tt.input)
			if got != tt.want {
				t.Errorf("convertSchedule(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
