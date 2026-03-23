package servicesstatus

import (
	"context"
	"time"

	domainservicesstatus "phant/internal/domain/servicesstatus"
)

type Dependencies struct {
	Now              func() time.Time
	Platform         func() string
	DiscoverServices func(context.Context) ([]domainservicesstatus.ServiceStatus, []string, error)
}

type Service struct {
	deps Dependencies
}

func NewService(deps Dependencies) *Service {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.Platform == nil {
		deps.Platform = func() string { return "unknown" }
	}

	return &Service{deps: deps}
}

func (s *Service) GetSnapshot(ctx context.Context) domainservicesstatus.Snapshot {
	snapshot := domainservicesstatus.Snapshot{
		GeneratedAt: s.deps.Now().UTC().Format(time.RFC3339),
		Platform:    s.deps.Platform(),
		Supported:   s.deps.Platform() == "linux",
	}

	if s.deps.DiscoverServices == nil {
		snapshot.Warnings = append(snapshot.Warnings, "service discovery is unavailable")
		return snapshot
	}

	services, warnings, err := s.deps.DiscoverServices(ctx)
	snapshot.Services = services
	snapshot.Warnings = append(snapshot.Warnings, warnings...)
	if err != nil {
		snapshot.LastError = err.Error()
	}

	if !snapshot.Supported {
		snapshot.Warnings = append(snapshot.Warnings, "Services status is currently supported on Linux only.")
	}

	return snapshot
}
