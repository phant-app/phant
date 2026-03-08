package main

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"phant/internal/services"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestGetRecentEvents_ReturnsEmptyWhenCollectorMissing(t *testing.T) {
	dumpService := services.NewAppServices().Dump
	got := dumpService.GetRecentEvents(10)
	if len(got) != 0 {
		t.Fatalf("dumpService.GetRecentEvents(10) len = %d, want %d", len(got), 0)
	}
}

func TestGetRecentEvents_ReturnsLatestN(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "collector.sock")
	appServices := services.NewAppServicesWithOptions(services.Options{SocketPath: socketPath})
	if err := appServices.Lifecycle.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatalf("Lifecycle.ServiceStartup() error = %v", err)
	}
	defer func() {
		_ = appServices.Lifecycle.ServiceShutdown()
	}()

	dumpService := appServices.Dump

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("net.Dial(unix, %q) error = %v", socketPath, err)
	}
	defer conn.Close()

	for _, id := range []string{"evt-1", "evt-2", "evt-3"} {
		if _, err := fmt.Fprintln(conn, validAppTestCLIEventLine(id)); err != nil {
			t.Fatalf("write event line for %q error = %v", id, err)
		}
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(dumpService.GetRecentEvents(0)) >= 3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	got := dumpService.GetRecentEvents(2)
	if len(got) != 2 {
		t.Fatalf("dumpService.GetRecentEvents(2) len = %d, want %d", len(got), 2)
	}
	if got[0].ID != "evt-2" || got[1].ID != "evt-3" {
		t.Fatalf("dumpService.GetRecentEvents(2) IDs = [%s %s], want [evt-2 evt-3]", got[0].ID, got[1].ID)
	}
}

func validAppTestCLIEventLine(id string) string {
	return fmt.Sprintf(`{"schemaVersion":1,"id":"%s","timestamp":"2026-03-02T12:00:00Z","sourceType":"cli","projectRoot":"/tmp/app","phpSapi":"cli","requestId":null,"command":{"name":"artisan"},"isDd":false,"payloadFormat":"json","payload":{"ok":true},"trace":[],"host":{"hostname":"test-host","pid":1234}}`, id)
}
