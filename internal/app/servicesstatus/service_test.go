package servicesstatus

import (
	"context"
	"errors"
	"testing"
	"time"

	domainservicesstatus "phant/internal/domain/servicesstatus"
)

func TestService_GetSnapshot(t *testing.T) {
	now := time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)
	svc := NewService(Dependencies{
		Now:      func() time.Time { return now },
		Platform: func() string { return "linux" },
		DiscoverServices: func(context.Context) ([]domainservicesstatus.ServiceStatus, []string, error) {
			return []domainservicesstatus.ServiceStatus{{
				ID:    "redis",
				Label: "Redis",
				State: domainservicesstatus.StateRunning,
			}}, []string{"sample warning"}, nil
		},
	})

	snapshot := svc.GetSnapshot(context.Background())
	if snapshot.GeneratedAt == "" {
		t.Fatalf("expected GeneratedAt")
	}
	if snapshot.Platform != "linux" {
		t.Fatalf("expected linux platform, got %s", snapshot.Platform)
	}
	if !snapshot.Supported {
		t.Fatalf("expected supported snapshot")
	}
	if len(snapshot.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(snapshot.Services))
	}
	if len(snapshot.Warnings) != 1 {
		t.Fatalf("expected warnings to include provider warning")
	}
}

func TestService_GetSnapshot_ErrorAndUnsupported(t *testing.T) {
	svc := NewService(Dependencies{
		Platform: func() string { return "darwin" },
		DiscoverServices: func(context.Context) ([]domainservicesstatus.ServiceStatus, []string, error) {
			return nil, nil, errors.New("boom")
		},
	})

	snapshot := svc.GetSnapshot(context.Background())
	if snapshot.Supported {
		t.Fatalf("expected unsupported snapshot")
	}
	if snapshot.LastError == "" {
		t.Fatalf("expected error message")
	}
	if len(snapshot.Warnings) == 0 {
		t.Fatalf("expected unsupported warning")
	}
}
