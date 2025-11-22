package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// Attribute represents a single S.M.A.R.T. attribute.
type Attribute struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Value  int    `json:"value"`
	Worst  int    `json:"worst"`
	Thresh int    `json:"thresh"`
	Raw    string `json:"raw"`
	Status string `json:"status"` // "OK" or "FAILING"
}

// Report represents a S.M.A.R.T. health report.
type Report struct {
	Disk       string      `json:"disk"`
	Passed     bool        `json:"passed"`
	Attributes []Attribute `json:"attributes"`
	CheckedAt  time.Time   `json:"checked_at"`
}

// smartctlOutput represents the JSON output from smartctl.
type smartctlOutput struct {
	SmartStatus struct {
		Passed bool `json:"passed"`
	} `json:"smart_status"`
	AtaSmartAttributes struct {
		Table []struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Value      int    `json:"value"`
			Worst      int    `json:"worst"`
			Thresh     int    `json:"thresh"`
			WhenFailed string `json:"when_failed"`
			Raw        struct {
				Value  int64  `json:"value"`
				String string `json:"string"`
			} `json:"raw"`
		} `json:"table"`
	} `json:"ata_smart_attributes"`
}

// Smart retrieves S.M.A.R.T. data for a disk using smartctl.
func Smart(ctx context.Context, name string) (*Report, error) {
	// On macOS, return mock data for development
	if runtime.GOOS == "darwin" {
		return smartMock(name), nil
	}

	// On Linux, use smartctl
	m := NewManager()
	return m.smartLinux(ctx, name)
}

// smartLinux executes smartctl and parses the output.
func (m *Manager) smartLinux(ctx context.Context, name string) (*Report, error) {
	// Execute smartctl with JSON output
	// -a: all information
	// -j: JSON format
	output, err := m.exec.CombinedOutput(ctx, "smartctl", "-a", "-j", "/dev/"+name)

	// smartctl returns non-zero exit code even on success in some cases
	// We check the JSON output instead
	if err != nil {
		// Check if it's just exit code issue
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit codes 0-3 are acceptable (device opened, smart enabled, etc)
			if exitErr.ExitCode() > 3 {
				return nil, fmt.Errorf("smartctl failed: %w", err)
			}
		} else {
			return nil, fmt.Errorf("smartctl execution failed: %w", err)
		}
	}

	var result smartctlOutput
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("parse smartctl output: %w", err)
	}

	// Convert to our Report format
	report := &Report{
		Disk:      name,
		Passed:    result.SmartStatus.Passed,
		CheckedAt: time.Now(),
	}

	// Convert attributes
	for _, attr := range result.AtaSmartAttributes.Table {
		status := "OK"
		if attr.WhenFailed != "" && attr.WhenFailed != "-" {
			status = "FAILING"
		}

		report.Attributes = append(report.Attributes, Attribute{
			ID:     attr.ID,
			Name:   attr.Name,
			Value:  attr.Value,
			Worst:  attr.Worst,
			Thresh: attr.Thresh,
			Raw:    attr.Raw.String,
			Status: status,
		})
	}

	return report, nil
}

// smartMock returns mock data for development (macOS).
func smartMock(name string) *Report {
	return &Report{
		Disk:      name,
		Passed:    true,
		CheckedAt: time.Now(),
		Attributes: []Attribute{
			{ID: 1, Name: "Raw_Read_Error_Rate", Value: 100, Worst: 100, Thresh: 51, Raw: "0", Status: "OK"},
			{ID: 5, Name: "Reallocated_Sector_Ct", Value: 100, Worst: 100, Thresh: 10, Raw: "0", Status: "OK"},
			{ID: 9, Name: "Power_On_Hours", Value: 99, Worst: 99, Thresh: 0, Raw: "1234", Status: "OK"},
			{ID: 194, Name: "Temperature_Celsius", Value: 64, Worst: 64, Thresh: 0, Raw: "36", Status: "OK"},
			{ID: 197, Name: "Current_Pending_Sector", Value: 100, Worst: 100, Thresh: 0, Raw: "0", Status: "OK"},
			{ID: 198, Name: "Offline_Uncorrectable", Value: 100, Worst: 100, Thresh: 0, Raw: "0", Status: "OK"},
		},
	}
}

// CheckHealth runs a S.M.A.R.T. check and returns an error if the disk is failing.
func CheckHealth(ctx context.Context, name string) error {
	report, err := Smart(ctx, name)
	if err != nil {
		return fmt.Errorf("smart check failed: %w", err)
	}

	if !report.Passed {
		return fmt.Errorf("disk %s is failing S.M.A.R.T. check", name)
	}

	return nil
}
