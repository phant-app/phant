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
	return "", fmt.Errorf("command not mocked: %s", key)
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
			"systemctl list-unit-files --type=service --no-legend --plain": "redis.service enabled\nmysql.service enabled",
			"systemctl is-active redis.service":                            "active",
			"ss -ltnpH":                                                    "",
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
	if lookup["redis"].Unit != "redis.service" {
		t.Fatalf("redis unit mismatch: expected redis.service, got %s", lookup["redis"].Unit)
	}
	if lookup["mysql"].State != servicesstatus.StateStopped {
		t.Fatalf("mysql state mismatch: expected stopped, got %s", lookup["mysql"].State)
	}
	if lookup["mysql"].Unit != "mysql.service" {
		t.Fatalf("mysql unit mismatch: expected mysql.service, got %s", lookup["mysql"].Unit)
	}
	if lookup["mailpit"].State != servicesstatus.StateUnavailable {
		t.Fatalf("mailpit state mismatch: expected unavailable, got %s", lookup["mailpit"].State)
	}
}

func TestProvider_DiscoverServices_UsesDetectedListeningPort(t *testing.T) {
	runner := mockRunner{
		outputs: map[string]string{
			"systemctl list-unit-files --type=service --no-legend --plain": "redis.service enabled",
			"systemctl is-active redis.service":                            "active",
			"ss -ltnpH":                                                    "LISTEN 0 511 127.0.0.1:6380 0.0.0.0:* users:((\"redis-server\",pid=111,fd=6))",
		},
	}

	provider := NewProvider(runner)
	services, warnings, err := provider.DiscoverServices(context.Background())
	if err != nil {
		t.Fatalf("DiscoverServices() error = %v, want nil", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("DiscoverServices() warnings = %v, want none", warnings)
	}

	lookup := map[string]servicesstatus.ServiceStatus{}
	for _, service := range services {
		lookup[service.ID] = service
	}

	if got := lookup["redis"].Port; got != 6380 {
		t.Fatalf("DiscoverServices(redis).Port = %d, want %d", got, 6380)
	}
}

func TestProvider_DiscoverServices_MissingUnitIsUnavailable(t *testing.T) {
	runner := mockRunner{
		outputs: map[string]string{
			"systemctl list-unit-files --type=service --no-legend --plain": "redis.service enabled",
			"systemctl is-active redis.service":                            "active",
			"ss -ltnpH":                                                    "",
		},
	}

	provider := NewProvider(runner)
	services, warnings, err := provider.DiscoverServices(context.Background())
	if err != nil {
		t.Fatalf("DiscoverServices() error = %v, want nil", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("DiscoverServices() warnings = %v, want none", warnings)
	}

	lookup := map[string]servicesstatus.ServiceStatus{}
	for _, service := range services {
		lookup[service.ID] = service
	}

	if got := lookup["mysql"].State; got != servicesstatus.StateUnavailable {
		t.Fatalf("DiscoverServices(mysql).State = %s, want %s", got, servicesstatus.StateUnavailable)
	}
	if got := lookup["redis"].State; got != servicesstatus.StateRunning {
		t.Fatalf("DiscoverServices(redis).State = %s, want %s", got, servicesstatus.StateRunning)
	}
}

func TestProvider_DiscoverServices_UsesDetectedUnitVariant(t *testing.T) {
	runner := mockRunner{
		outputs: map[string]string{
			"systemctl list-unit-files --type=service --no-legend --plain": "valkey.1.service enabled",
			"systemctl is-active valkey.1.service":                         "active",
			"ss -ltnpH":                                                    "LISTEN 0 511 127.0.0.1:6390 0.0.0.0:* users:((\"valkey-server\",pid=111,fd=6))",
		},
	}

	provider := NewProvider(runner)
	services, warnings, err := provider.DiscoverServices(context.Background())
	if err != nil {
		t.Fatalf("DiscoverServices() error = %v, want nil", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("DiscoverServices() warnings = %v, want none", warnings)
	}

	lookup := map[string]servicesstatus.ServiceStatus{}
	for _, service := range services {
		lookup[service.ID] = service
	}

	if got := lookup["valkey"].Unit; got != "valkey.1.service" {
		t.Fatalf("DiscoverServices(valkey).Unit = %s, want %s", got, "valkey.1.service")
	}
	if got := lookup["valkey"].State; got != servicesstatus.StateRunning {
		t.Fatalf("DiscoverServices(valkey).State = %s, want %s", got, servicesstatus.StateRunning)
	}
	if got := lookup["valkey"].Port; got != 6390 {
		t.Fatalf("DiscoverServices(valkey).Port = %d, want %d", got, 6390)
	}
}
