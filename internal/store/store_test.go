package store

import (
	"reflect"
	"testing"
	"time"
)

func TestMonitorStore_UpdateAndGetAll(t *testing.T) {
	type update struct {
		id       string
		cpu, ram float64
	}

	type testCase struct {
		name    string
		updates []update
		want    map[string]NodeStats
	}

	testCases := []testCase{
		{
			name: "single node",
			updates: []update{
				{id: "node-1", cpu: 50.5, ram: 75.2},
			},
			want: map[string]NodeStats{
				"node-1": {CPU: 50.5, RAM: 75.2},
			},
		},
		{
			name: "multiple nodes",
			updates: []update{
				{id: "node-1", cpu: 10.0, ram: 20.0},
				{id: "node-2", cpu: 30.0, ram: 40.0},
			},
			want: map[string]NodeStats{
				"node-1": {CPU: 10.0, RAM: 20.0},
				"node-2": {CPU: 30.0, RAM: 40.0},
			},
		},
		{
			name: "update existing node",
			updates: []update{
				{id: "node-1", cpu: 10.0, ram: 20.0},
				{id: "node-1", cpu: 99.9, ram: 88.8},
			},
			want: map[string]NodeStats{
				"node-1": {CPU: 99.9, RAM: 88.8},
			},
		},
		{
			name:    "no nodes",
			updates: []update{},
			want:    map[string]NodeStats{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewMonitorStore()

			for _, u := range tc.updates {
				store.Update(u.id, u.cpu, u.ram)
			}

			got := store.GetAll()

			for id, stats := range got {
				if time.Since(stats.LastSeen) > 5*time.Second {
					t.Errorf("node '%s' LastSeen is too old: %v", id, stats.LastSeen)
				}

				stats.LastSeen = time.Time{}
				got[id] = stats

				if wantedStats, ok := tc.want[id]; ok {
					wantedStats.LastSeen = time.Time{}
					tc.want[id] = wantedStats
				}
			}

			if len(got) != len(tc.want) {
				t.Fatalf("got %d nodes, but want %d", len(got), len(tc.want))
			}

			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("got %v stats, but want %v", got, tc.want)
			}
		})
	}
}
