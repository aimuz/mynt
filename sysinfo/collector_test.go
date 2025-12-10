package sysinfo

import (
	"testing"
	"time"
)

// BenchmarkListProcesses benchmarks the optimized procfs-based implementation.
func BenchmarkListProcesses(b *testing.B) {
	c := NewCollector()
	// Warmup: ensure lastCPU is initialized
	if _, err := c.ListProcesses(); err != nil {
		b.Fatalf("warmup failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = c.ListProcesses()
	}
}
