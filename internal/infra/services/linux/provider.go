package linux

import (
	"context"
	"sort"
	"strings"

	"phant/internal/domain/servicesstatus"
	"phant/internal/infra/system"
)

type serviceDefinition struct {
	ID          string
	Label       string
	Description string
	Port        int
	Unit        string
}

var defaultServices = []serviceDefinition{
	{ID: "postgresql", Label: "PostgreSQL", Description: "PostgreSQL database server", Port: 5432, Unit: "postgresql.service"},
	{ID: "mysql", Label: "MySQL", Description: "MySQL database server", Port: 3306, Unit: "mysql.service"},
	{ID: "mariadb", Label: "MariaDB", Description: "MariaDB database server", Port: 3306, Unit: "mariadb.service"},
	{ID: "valkey", Label: "Valkey", Description: "Valkey key-value store", Port: 6379, Unit: "valkey.service"},
	{ID: "redis", Label: "Redis", Description: "Redis key-value store", Port: 6379, Unit: "redis.service"},
	{ID: "mailpit", Label: "Mailpit", Description: "Mailpit email testing server", Port: 1025, Unit: "mailpit.service"},
}

type Provider struct {
	runner system.Runner
}

func NewProvider(runner system.Runner) *Provider {
	return &Provider{runner: runner}
}

func (p *Provider) Platform() string {
	return p.runner.GOOS()
}

func (p *Provider) DiscoverServices(ctx context.Context) ([]servicesstatus.ServiceStatus, []string, error) {
	_, err := p.runner.LookPath("systemctl")
	if err != nil {
		statuses := make([]servicesstatus.ServiceStatus, 0, len(defaultServices))
		for _, def := range defaultServices {
			statuses = append(statuses, servicesstatus.ServiceStatus{
				ID:          def.ID,
				Label:       def.Label,
				Description: def.Description,
				Port:        def.Port,
				Unit:        def.Unit,
				State:       servicesstatus.StateUnavailable,
			})
		}
		sortStatuses(statuses)
		return statuses, []string{"systemctl is unavailable on this machine"}, nil
	}

	statuses := make([]servicesstatus.ServiceStatus, 0, len(defaultServices))
	for _, def := range defaultServices {
		statuses = append(statuses, p.inspectService(ctx, def))
	}

	sortStatuses(statuses)
	return statuses, nil, nil
}

func (p *Provider) inspectService(ctx context.Context, def serviceDefinition) servicesstatus.ServiceStatus {
	status := servicesstatus.ServiceStatus{
		ID:          def.ID,
		Label:       def.Label,
		Description: def.Description,
		Port:        def.Port,
		Unit:        def.Unit,
	}

	if !p.unitExists(ctx, def.Unit) {
		status.State = servicesstatus.StateUnavailable
		return status
	}

	activeState, err := p.readActiveState(ctx, def.Unit)
	if err != nil {
		status.State = servicesstatus.StateStopped
		return status
	}

	if activeState == "active" {
		status.State = servicesstatus.StateRunning
		return status
	}

	status.State = servicesstatus.StateStopped
	return status
}

func (p *Provider) unitExists(ctx context.Context, unit string) bool {
	stdout, err := p.runner.Run(ctx, "systemctl", "list-unit-files", "--type=service", "--no-legend", "--plain", unit)
	if err != nil {
		return false
	}

	line := strings.TrimSpace(stdout)
	if line == "" {
		return false
	}

	return strings.Contains(line, unit)
}

func (p *Provider) readActiveState(ctx context.Context, unit string) (string, error) {
	stdout, err := p.runner.Run(ctx, "systemctl", "is-active", unit)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

func sortStatuses(statuses []servicesstatus.ServiceStatus) {
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Label < statuses[j].Label
	})
}
