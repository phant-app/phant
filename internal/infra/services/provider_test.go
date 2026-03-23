package services

import (
	"context"
	"testing"
)

func TestUnsupportedProvider_DiscoverServices_EmptySlice(t *testing.T) {
	provider := newUnsupportedProvider("darwin")

	services, warnings, err := provider.DiscoverServices(context.Background())
	if err != nil {
		t.Fatalf("DiscoverServices() error = %v, want nil", err)
	}
	if services == nil {
		t.Fatalf("DiscoverServices() services = nil, want empty slice")
	}
	if len(services) != 0 {
		t.Fatalf("DiscoverServices() len(services) = %d, want 0", len(services))
	}
	if len(warnings) == 0 {
		t.Fatalf("DiscoverServices() warnings = none, want at least one warning")
	}
}
