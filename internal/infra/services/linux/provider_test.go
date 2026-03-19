package linux

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"phant/internal/domain/servicesstatus"
)

type mockRunner struct {
	lookPathErr error
	outputs     map[string]string
	errors      map[string]error
}

func (m mockRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	key := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	if err, ok := m.errors[key]; ok {
		return "", err
	}
	if output, ok := m.outputs[key]; ok {
		return output, nil
	}
	return "", errors.New("command not mocked")
}

func (m mockRunner) LookPath(_ string) (string, error) {
	if m.lookPathErr != nil {
		return "", m.lookPathErr
	}
	return "/usr/bin/systemctl", nil
}

func (m mockRunner) GOOS() string {
	return "linux"
}

func TestProvider_DiscoverServices_SystemctlUnavailable(t *testing.T) {
	provider := NewProvider(mockRunner{lookPathErr: errors.New("missing")})

	services, warnings, err := provider.DiscoverServices(context.Background())
	if err != nil {
		t.Fatalf("DiscoverServices returned error: %v", err)
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warning when systemctl unavailable")
	}
	if len(services) != len(defaultServices) {
		t.Fatalf("expected %d services, got %d", len(defaultServices), len(services))
	}
	for _, service := range services {
		if service.State != servicesstatus.StateUnavailable {
			t.Fatalf("expected unavailable state, got %s for %s", service.State, service.ID)
		}
	}
}

func TestProvider_DiscoverServices_RunningStoppedUnavailable(t *testing.T) {
	runner := mockRunner{
		outputs: map[string]string{
			"systemctl list-unit-files --type=service --no-legend --plain redis.service": "redis.service enabled",
			"systemctl is-active redis.service":                                          "active",
			"systemctl list-unit-files --type=service --no-legend --plain mysql.service": "mysql.service enabled",
		},
		errors: map[string]error{
			"systemctl is-active mysql.service": errors.New("inactive"),
		},
	}
	provider := NewProvider(runner)

	services, warnings, err := provider.DiscoverServices(context.Background())
	if err != nil {
		t.Fatalf("DiscoverServices returned error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}

	lookup := map[string]servicesstatus.ServiceStatus{}
	for _, service := range services {
		lookup[service.ID] = service
	}

	if lookup["redis"].State != servicesstatus.StateRunning {
		t.Fatalf("redis state mismatch: expected running, got %s", lookup["redis"].State)
	}
	if lookup["mysql"].State != servicesstatus.StateStopped {
		t.Fatalf("mysql state mismatch: expected stopped, got %s", lookup["mysql"].State)
	}
	if lookup["mailpit"].State != servicesstatus.StateUnavailable {
		t.Fatalf("mailpit state mismatch: expected unavailable, got %s", lookup["mailpit"].State)
	}
}
