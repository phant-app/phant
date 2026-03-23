package linux

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"phant/internal/domain/servicesstatus"
	"phant/internal/infra/system"
)

type serviceDefinition struct {
	ID           string
	Label        string
	Description  string
	Port         int
	UnitNames    []string
	UnitPrefixes []string
	Processes    []string
}

var defaultServices = []serviceDefinition{
	{ID: "postgresql", Label: "PostgreSQL", Description: "PostgreSQL database server", Port: 5432, UnitNames: []string{"postgresql.service"}, UnitPrefixes: []string{"postgresql@", "postgresql."}, Processes: []string{"postgres"}},
	{ID: "mysql", Label: "MySQL", Description: "MySQL database server", Port: 3306, UnitNames: []string{"mysql.service", "mysqld.service"}, UnitPrefixes: []string{"mysql@", "mysql.", "mysqld@", "mysqld."}, Processes: []string{"mysqld", "mariadbd"}},
	{ID: "mariadb", Label: "MariaDB", Description: "MariaDB database server", Port: 3306, UnitNames: []string{"mariadb.service"}, UnitPrefixes: []string{"mariadb@", "mariadb."}, Processes: []string{"mariadbd", "mysqld"}},
	{ID: "valkey", Label: "Valkey", Description: "Valkey key-value store", Port: 6379, UnitNames: []string{"valkey.service"}, UnitPrefixes: []string{"valkey@", "valkey."}, Processes: []string{"valkey-server"}},
	{ID: "redis", Label: "Redis", Description: "Redis key-value store", Port: 6379, UnitNames: []string{"redis.service", "redis-server.service"}, UnitPrefixes: []string{"redis@", "redis.", "redis-server@", "redis-server."}, Processes: []string{"redis-server"}},
	{ID: "mailpit", Label: "Mailpit", Description: "Mailpit email testing server", Port: 1025, UnitNames: []string{"mailpit.service"}, UnitPrefixes: []string{"mailpit@", "mailpit."}, Processes: []string{"mailpit"}},
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
				State:       servicesstatus.StateUnavailable,
			})
		}
		sortStatuses(statuses)
		return statuses, []string{"systemctl is unavailable on this machine"}, nil
	}

	installedUnits, err := p.readInstalledUnits(ctx)
	if err != nil {
		return nil, nil, err
	}

	statuses := make([]servicesstatus.ServiceStatus, 0, len(defaultServices))
	portsByProcess := p.readListeningPortsByProcess(ctx)
	for _, def := range defaultServices {
		statuses = append(statuses, p.inspectService(ctx, def, installedUnits, portsByProcess))
	}

	sortStatuses(statuses)
	return statuses, nil, nil
}

func (p *Provider) inspectService(
	ctx context.Context,
	def serviceDefinition,
	installedUnits map[string]struct{},
	portsByProcess map[string]int,
) servicesstatus.ServiceStatus {
	status := servicesstatus.ServiceStatus{
		ID:          def.ID,
		Label:       def.Label,
		Description: def.Description,
		Port:        def.Port,
	}

	unitName, exists := resolveInstalledUnit(def, installedUnits)
	if !exists {
		status.State = servicesstatus.StateUnavailable
		return status
	}
	status.Unit = unitName

	activeState, err := p.readActiveState(ctx, unitName)
	if err != nil {
		status.State = servicesstatus.StateStopped
		return status
	}

	if activeState == "active" {
		status.State = servicesstatus.StateRunning
		if detectedPort, ok := detectPort(def.Processes, portsByProcess); ok {
			status.Port = detectedPort
		}
		return status
	}

	status.State = servicesstatus.StateStopped
	return status
}

func (p *Provider) readInstalledUnits(ctx context.Context) (map[string]struct{}, error) {
	stdout, err := p.runner.Run(ctx, "systemctl", "list-unit-files", "--type=service", "--no-legend", "--plain")
	if err != nil {
		return nil, err
	}

	units := make(map[string]struct{})
	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		unitName := strings.TrimSpace(fields[0])
		if !strings.HasSuffix(unitName, ".service") {
			continue
		}
		units[unitName] = struct{}{}
	}

	return units, nil
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

func (p *Provider) readListeningPortsByProcess(ctx context.Context) map[string]int {
	stdout, err := p.runner.Run(ctx, "ss", "-ltnpH")
	if err != nil || strings.TrimSpace(stdout) == "" {
		return map[string]int{}
	}

	portsByProcess := map[string]int{}
	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		localAddress := fields[3]
		port, ok := parsePort(localAddress)
		if !ok {
			continue
		}

		processField := strings.Join(fields[5:], " ")
		for _, process := range extractProcessNames(processField) {
			if _, exists := portsByProcess[process]; !exists {
				portsByProcess[process] = port
			}
		}
	}

	return portsByProcess
}

func detectPort(processes []string, portsByProcess map[string]int) (int, bool) {
	for _, process := range processes {
		if port, ok := portsByProcess[process]; ok {
			return port, true
		}
	}

	return 0, false
}

func parsePort(localAddress string) (int, bool) {
	separator := strings.LastIndex(localAddress, ":")
	if separator == -1 || separator == len(localAddress)-1 {
		return 0, false
	}

	portText := localAddress[separator+1:]
	if portText == "*" {
		return 0, false
	}

	port, err := strconv.Atoi(portText)
	if err != nil {
		return 0, false
	}

	return port, true
}

func extractProcessNames(processField string) []string {
	names := make([]string, 0)
	segments := strings.Split(processField, "\"")
	for i := 1; i < len(segments); i += 2 {
		name := strings.TrimSpace(segments[i])
		if name == "" {
			continue
		}
		names = append(names, name)
	}

	return names
}

func resolveInstalledUnit(def serviceDefinition, installedUnits map[string]struct{}) (string, bool) {
	for _, unitName := range def.UnitNames {
		if _, ok := installedUnits[unitName]; ok {
			return unitName, true
		}
	}

	candidates := make([]string, 0)
	for installedUnit := range installedUnits {
		if !strings.HasSuffix(installedUnit, ".service") {
			continue
		}
		if hasAnyPrefix(installedUnit, def.UnitPrefixes) {
			candidates = append(candidates, installedUnit)
		}
	}

	if len(candidates) == 0 {
		return "", false
	}

	sort.Strings(candidates)
	return candidates[0], true
}

func hasAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}

	return false
}
