package runtime

import "testing"

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.HTTPPort != 8080 {
		t.Errorf("HTTPPort = %d, want 8080", opts.HTTPPort)
	}
	if opts.GRPCPort != 9090 {
		t.Errorf("GRPCPort = %d, want 9090", opts.GRPCPort)
	}
}
