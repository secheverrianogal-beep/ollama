package discover

import (
	"testing"
	"time"
)

func TestGPUMemoryRefreshTimeoutScalesWithDeviceCount(t *testing.T) {
	tests := []struct {
		name        string
		deviceCount int
		want        time.Duration
	}{
		{name: "no devices", deviceCount: 0, want: 3 * time.Second},
		{name: "one device uses floor", deviceCount: 1, want: 3 * time.Second},
		{name: "three devices use floor", deviceCount: 3, want: 3 * time.Second},
		{name: "six devices scale timeout", deviceCount: 6, want: 6 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gpuMemoryRefreshTimeout(tt.deviceCount); got != tt.want {
				t.Fatalf("gpuMemoryRefreshTimeout(%d) = %s, want %s", tt.deviceCount, got, tt.want)
			}
		})
	}
}
