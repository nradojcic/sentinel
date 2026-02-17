package collector

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type Stats struct {
	CPU float64
	RAM float64
}

// StatsCollector defines an interface for collecting system stats
// This allows for mocking in tests
type StatsCollector interface {
	CPUPercent() ([]float64, error)
	VirtualMemory() (*mem.VirtualMemoryStat, error)
}

// GopsutilCollector is a real implementation of StatsCollector using gopsutil
type GopsutilCollector struct{}

// CPUPercent implements the StatsCollector interface for CPU stats
func (c GopsutilCollector) CPUPercent() ([]float64, error) {
	// For 0 duration, it reads overall CPU usage since last call or system boot
	return cpu.Percent(time.Duration(0), false)
}

// VirtualMemory implements the StatsCollector interface for memory stats
func (c GopsutilCollector) VirtualMemory() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

// GetStats collects current CPU and RAM usage using the real gopsutil collector
// The public-facing function
func GetStats() (Stats, error) {
	return getStats(GopsutilCollector{})
}

// getStats is the internal, testable function that contains the core logic
func getStats(collector StatsCollector) (Stats, error) {
	cpuPercent, err := collector.CPUPercent()
	if err != nil {
		return Stats{}, fmt.Errorf("failed to get CPU percent: %w", err)
	}
	if len(cpuPercent) == 0 {
		return Stats{}, fmt.Errorf("cpu.Percent returned no values")
	}

	vMem, err := collector.VirtualMemory()
	if err != nil {
		return Stats{}, fmt.Errorf("failed to get virtual memory: %w", err)
	}

	return Stats{
		CPU: cpuPercent[0],
		RAM: vMem.UsedPercent,
	}, nil
}
