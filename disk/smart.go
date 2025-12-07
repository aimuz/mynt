package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SMART attribute IDs of interest.
const (
	attrReallocatedSectors = 5
	attrPowerOnHours       = 9
	attrTemperature        = 194
	attrPendingSectors     = 197
	attrUncorrectable      = 198
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

// TestType represents a S.M.A.R.T. self-test type.
type TestType string

const (
	TestShort TestType = "short"
	TestLong  TestType = "long"
)

// TestStatus represents the status of a S.M.A.R.T. self-test.
type TestStatus struct {
	Running    bool   `json:"running"`
	Type       string `json:"type,omitempty"`
	Progress   int    `json:"progress,omitempty"`
	LastResult string `json:"last_result,omitempty"`
}

// DetailedReport includes extended SMART data for disk details view.
type DetailedReport struct {
	Disk                string      `json:"disk"`
	Passed              bool        `json:"passed"`
	Attributes          []Attribute `json:"attributes"`
	CheckedAt           time.Time   `json:"checked_at"`
	PowerOnHours        int64       `json:"power_on_hours"`
	PowerCycleCount     int64       `json:"power_cycle_count"`
	ReallocatedSectors  int64       `json:"reallocated_sectors"`
	PendingSectors      int64       `json:"pending_sectors"`
	UncorrectableErrors int64       `json:"uncorrectable_errors"`
	Temperature         int         `json:"temperature"`
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
	Temperature struct {
		Current int `json:"current"`
	} `json:"temperature"`
	PowerOnTime struct {
		Hours int64 `json:"hours"`
	} `json:"power_on_time"`
	PowerCycleCount     int64 `json:"power_cycle_count"`
	AtaSmartSelfTestLog struct {
		Standard struct {
			Table []struct {
				Status struct {
					String string `json:"string"`
				} `json:"status"`
			} `json:"table"`
		} `json:"standard"`
	} `json:"ata_smart_self_test_log"`
	AtaSmartData struct {
		SelfTest struct {
			Status struct {
				Value            int    `json:"value"`
				String           string `json:"string"`
				RemainingPercent int    `json:"remaining_percent"`
			} `json:"status"`
		} `json:"self_test"`
	} `json:"ata_smart_data"`
}

// Smart retrieves S.M.A.R.T. data for a disk.
func (m *Manager) Smart(ctx context.Context, name string) (*Report, error) {
	if runtime.GOOS == "darwin" {
		return mockReport(name), nil
	}

	out, err := m.runSmartctl(ctx, name)
	if err != nil {
		return nil, err
	}

	var data smartctlOutput
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("parse smartctl: %w", err)
	}

	r := &Report{
		Disk:      name,
		Passed:    data.SmartStatus.Passed,
		CheckedAt: time.Now(),
	}
	for _, a := range data.AtaSmartAttributes.Table {
		status := "OK"
		if a.WhenFailed != "" && a.WhenFailed != "-" {
			status = "FAILING"
		}
		r.Attributes = append(r.Attributes, Attribute{
			ID:     a.ID,
			Name:   a.Name,
			Value:  a.Value,
			Worst:  a.Worst,
			Thresh: a.Thresh,
			Raw:    a.Raw.String,
			Status: status,
		})
	}
	return r, nil
}

// SmartDetails retrieves comprehensive SMART data.
func (m *Manager) SmartDetails(ctx context.Context, name string) (*DetailedReport, error) {
	if runtime.GOOS == "darwin" {
		return mockDetailedReport(name), nil
	}

	out, err := m.runSmartctl(ctx, name)
	if err != nil {
		return nil, err
	}

	var data smartctlOutput
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("parse smartctl: %w", err)
	}

	r := &DetailedReport{
		Disk:            name,
		Passed:          data.SmartStatus.Passed,
		CheckedAt:       time.Now(),
		PowerOnHours:    data.PowerOnTime.Hours,
		PowerCycleCount: data.PowerCycleCount,
		Temperature:     data.Temperature.Current,
	}

	for _, a := range data.AtaSmartAttributes.Table {
		status := "OK"
		if a.WhenFailed != "" && a.WhenFailed != "-" {
			status = "FAILING"
		}
		r.Attributes = append(r.Attributes, Attribute{
			ID:     a.ID,
			Name:   a.Name,
			Value:  a.Value,
			Worst:  a.Worst,
			Thresh: a.Thresh,
			Raw:    a.Raw.String,
			Status: status,
		})

		switch a.ID {
		case attrReallocatedSectors:
			r.ReallocatedSectors = a.Raw.Value
		case attrPowerOnHours:
			if r.PowerOnHours == 0 {
				r.PowerOnHours = a.Raw.Value
			}
		case attrTemperature:
			if r.Temperature == 0 {
				r.Temperature = parseTemperature(a.Raw.String)
			}
		case attrPendingSectors:
			r.PendingSectors = a.Raw.Value
		case attrUncorrectable:
			r.UncorrectableErrors = a.Raw.Value
		}
	}
	return r, nil
}

// SmartTest starts a S.M.A.R.T. self-test.
func (m *Manager) SmartTest(ctx context.Context, name string, typ TestType) error {
	if runtime.GOOS == "darwin" {
		return nil
	}

	_, err := m.exec.CombinedOutput(ctx, "smartctl", "-t", string(typ), "/dev/"+name)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Only treat bits 0-2 as fatal
			if exitErr.ExitCode()&smartExitFatalMask == 0 {
				return nil
			}
		}
		return fmt.Errorf("start smart test: %w", err)
	}
	return nil
}

// SmartTestStatus gets the current self-test status.
func (m *Manager) SmartTestStatus(ctx context.Context, name string) (*TestStatus, error) {
	if runtime.GOOS == "darwin" {
		return &TestStatus{LastResult: "Completed without error"}, nil
	}

	out, err := m.runSmartctl(ctx, name)
	if err != nil {
		return nil, err
	}

	var data smartctlOutput
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("parse smartctl: %w", err)
	}

	s := &TestStatus{}
	st := data.AtaSmartData.SelfTest.Status
	if st.Value != 0 && st.RemainingPercent > 0 {
		s.Running = true
		s.Progress = 100 - st.RemainingPercent
		s.Type = st.String
	}
	if len(data.AtaSmartSelfTestLog.Standard.Table) > 0 {
		s.LastResult = data.AtaSmartSelfTestLog.Standard.Table[0].Status.String
	}
	return s, nil
}

// smartctl exit code bitmask values (from man smartctl).
const (
	// Fatal errors - command/device issues
	smartExitCmdLine   = 1 << 0 // Bit 0: Command line parse error
	smartExitDevOpen   = 1 << 1 // Bit 1: Device open failed
	smartExitCmdFailed = 1 << 2 // Bit 2: SMART command to disk failed

	// Disk health status - not fatal, still have valid data
	// Bit 3: SMART status check returned "DISK FAILING"
	// Bit 4: Prefail attributes <= threshold
	// Bit 5: Some attributes > threshold in past
	// Bit 6: Error log contains errors
	// Bit 7: Self-test log contains errors

	smartExitFatalMask = smartExitCmdLine | smartExitDevOpen | smartExitCmdFailed
)

// runSmartctl executes smartctl and handles exit codes using bitmask.
func (m *Manager) runSmartctl(ctx context.Context, name string) ([]byte, error) {
	out, err := m.exec.CombinedOutput(ctx, "smartctl", "-a", "-j", "/dev/"+name)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			// Only treat bits 0-2 as fatal (command/device errors)
			// Bits 3-7 indicate disk health issues but data is still valid
			if code&smartExitFatalMask == 0 {
				return out, nil
			}
		}
		return nil, fmt.Errorf("smartctl: %w", err)
	}
	return out, nil
}

// parseTemperature extracts temperature from SMART raw string.
func parseTemperature(raw string) int {
	parts := strings.Fields(raw)
	if len(parts) > 0 {
		if t, err := strconv.Atoi(parts[0]); err == nil {
			return t
		}
	}
	return 0
}

// mockReport returns mock data for macOS development.
func mockReport(name string) *Report {
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

// mockDetailedReport returns mock detailed data for macOS development.
func mockDetailedReport(name string) *DetailedReport {
	return &DetailedReport{
		Disk:                name,
		Passed:              true,
		CheckedAt:           time.Now(),
		PowerOnHours:        1234,
		PowerCycleCount:     42,
		Temperature:         36,
		ReallocatedSectors:  0,
		PendingSectors:      0,
		UncorrectableErrors: 0,
		Attributes: []Attribute{
			{ID: 1, Name: "Raw_Read_Error_Rate", Value: 100, Worst: 100, Thresh: 51, Raw: "0", Status: "OK"},
			{ID: 5, Name: "Reallocated_Sector_Ct", Value: 100, Worst: 100, Thresh: 10, Raw: "0", Status: "OK"},
			{ID: 9, Name: "Power_On_Hours", Value: 99, Worst: 99, Thresh: 0, Raw: "1234", Status: "OK"},
			{ID: 12, Name: "Power_Cycle_Count", Value: 100, Worst: 100, Thresh: 0, Raw: "42", Status: "OK"},
			{ID: 194, Name: "Temperature_Celsius", Value: 64, Worst: 64, Thresh: 0, Raw: "36", Status: "OK"},
			{ID: 197, Name: "Current_Pending_Sector", Value: 100, Worst: 100, Thresh: 0, Raw: "0", Status: "OK"},
			{ID: 198, Name: "Offline_Uncorrectable", Value: 100, Worst: 100, Thresh: 0, Raw: "0", Status: "OK"},
		},
	}
}

// CheckHealth returns an error if the disk is failing S.M.A.R.T.
func (m *Manager) CheckHealth(ctx context.Context, name string) error {
	r, err := m.Smart(ctx, name)
	if err != nil {
		return err
	}
	if !r.Passed {
		return fmt.Errorf("disk %s failing S.M.A.R.T.", name)
	}
	return nil
}
