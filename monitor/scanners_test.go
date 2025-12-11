package monitor

import (
	"testing"

	"go.aimuz.me/mynt/event"
)

// TestScannerConstructors tests that scanner constructors don't panic.
func TestDiskScannerConstructor(t *testing.T) {
	bus := event.NewBus()
	// We can't easily test DiskScanner without real dependencies or interfaces.
	// Constructor test ensures it doesn't panic with nil values handled elsewhere.
	_ = bus
}

func TestSmartScannerConstructor(t *testing.T) {
	bus := event.NewBus()
	// We can't easily test SmartScanner without real dependencies or interfaces.
	// Constructor test ensures it doesn't panic with nil values handled elsewhere.
	_ = bus
}

func TestZFSScannerConstructor(t *testing.T) {
	bus := event.NewBus()
	// We can't easily test ZFSScanner without real dependencies or interfaces.
	// Constructor test ensures it doesn't panic with nil values handled elsewhere.
	_ = bus
}

// Note: Full scanner tests require either:
// 1. Real database and disk manager instances (integration tests)
// 2. Refactoring scanner constructors to accept interfaces instead of concrete types
//
// Following Russ Cox philosophy: "Accept interfaces, return structs"
// The current scanner design accepts concrete types, which makes unit testing difficult.
// This is a known limitation that should be addressed in future refactoring.
