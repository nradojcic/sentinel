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

// GetStats collects current CPU and RAM usage
func GetStats() (Stats, error) {
	// Get CPU usage over a short interval (e.g., 100ms)
	// For 0 duration, it reads overall CPU usage since last call or system boot
	cpuPercent, err := cpu.Percent(time.Duration(0), false)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to get CPU percent: %w", err)
	}

	// Get Virtual Memory stats
	vMem, err := mem.VirtualMemory()
	if err != nil {
		return Stats{}, fmt.Errorf("failed to get virtual memory: %w", err)
	}

	return Stats{
		CPU: cpuPercent[0],
		RAM: vMem.UsedPercent,
	}, nil
}
