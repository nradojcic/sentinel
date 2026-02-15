package collector

import (
	"errors"
	"reflect"
	"testing"

	"github.com/shirou/gopsutil/v3/mem"
)

// mockStatsCollector is a mock implementation of the StatsCollector interface for testing.
type mockStatsCollector struct {
	cpuPercent    []float64
	cpuErr        error
	virtualMemory *mem.VirtualMemoryStat
	memErr        error
}

func (m *mockStatsCollector) CPUPercent() ([]float64, error) {
	return m.cpuPercent, m.cpuErr
}

func (m *mockStatsCollector) VirtualMemory() (*mem.VirtualMemoryStat, error) {
	return m.virtualMemory, m.memErr
}

func TestGetStats(t *testing.T) {
	testCases := []struct {
		name      string
		collector StatsCollector
		want      Stats
		wantErr   bool
	}{
		{
			name: "success",
			collector: &mockStatsCollector{
				cpuPercent: []float64{25.5},
				virtualMemory: &mem.VirtualMemoryStat{
					UsedPercent: 60.1,
				},
			},
			want:    Stats{CPU: 25.5, RAM: 60.1},
			wantErr: false,
		},
		{
			name: "cpu error",
			collector: &mockStatsCollector{
				cpuErr: errors.New("cpu failed"),
			},
			want:    Stats{},
			wantErr: true,
		},
		{
			name: "memory error",
			collector: &mockStatsCollector{
				cpuPercent: []float64{10.0},
				memErr:     errors.New("mem failed"),
			},
			want:    Stats{},
			wantErr: true,
		},
		{
			name: "cpu percent returns empty slice",
			collector: &mockStatsCollector{
				cpuPercent: []float64{},
				virtualMemory: &mem.VirtualMemoryStat{
					UsedPercent: 50.0,
				},
			},
			want:    Stats{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getStats(tc.collector)

			if (err != nil) != tc.wantErr {
				t.Fatalf("getStats() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getStats() = %v, want %v", got, tc.want)
			}
		})
	}
}
