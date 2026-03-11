package health

import (
	"testing"
	"time"
)

func TestOptions_applyDefaults(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want Options
	}{
		{
			name: "all defaults",
			opts: Options{},
			want: Options{
				ListenAddr:      ":8080",
				StaleThreshold:  5 * time.Minute,
				ReadTimeout:     5 * time.Second,
				WriteTimeout:    10 * time.Second,
				ShutdownTimeout: 10 * time.Second,
			},
		},
		{
			name: "custom values preserved",
			opts: Options{
				ListenAddr:      ":9090",
				StaleThreshold:  10 * time.Minute,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    15 * time.Second,
				ShutdownTimeout: 20 * time.Second,
			},
			want: Options{
				ListenAddr:      ":9090",
				StaleThreshold:  10 * time.Minute,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    15 * time.Second,
				ShutdownTimeout: 20 * time.Second,
			},
		},
		{
			name: "partial custom values",
			opts: Options{
				ListenAddr:     ":3000",
				StaleThreshold: 2 * time.Minute,
			},
			want: Options{
				ListenAddr:      ":3000",
				StaleThreshold:  2 * time.Minute,
				ReadTimeout:     5 * time.Second,
				WriteTimeout:    10 * time.Second,
				ShutdownTimeout: 10 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.applyDefaults()
			if got != tt.want {
				t.Errorf("applyDefaults() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
