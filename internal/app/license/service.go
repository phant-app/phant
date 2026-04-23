package license

import (
	"context"
	"strings"

	domainlicense "phant/internal/domain/license"
	domainsettings "phant/internal/domain/settings"
)

type Dependencies struct {
	LoadSettings func(context.Context) (domainsettings.AppSettings, error)
	SaveSettings func(context.Context, domainsettings.AppSettings) error
}

type Service struct {
	deps Dependencies
}

func NewService(deps Dependencies) *Service {
	return &Service{deps: deps}
}

func (s *Service) GetKey(ctx context.Context) domainlicense.KeyResult {
	if s.deps.LoadSettings == nil {
		return domainlicense.KeyResult{Error: "settings read operation is unavailable"}
	}

	settings, err := s.deps.LoadSettings(ctx)
	if err != nil {
		return domainlicense.KeyResult{Error: err.Error()}
	}

	return domainlicense.KeyResult{LicenseKey: strings.TrimSpace(settings.LicenseKey)}
}

func (s *Service) SaveKey(ctx context.Context, licenseKey string) domainlicense.SaveResult {
	if s.deps.LoadSettings == nil || s.deps.SaveSettings == nil {
		return domainlicense.SaveResult{Error: "settings write operation is unavailable"}
	}

	settings, err := s.deps.LoadSettings(ctx)
	if err != nil {
		return domainlicense.SaveResult{Error: err.Error()}
	}

	normalized := strings.TrimSpace(licenseKey)
	settings.LicenseKey = normalized
	if err := s.deps.SaveSettings(ctx, settings); err != nil {
		return domainlicense.SaveResult{Error: err.Error()}
	}

	if normalized == "" {
		return domainlicense.SaveResult{
			Success:    true,
			LicenseKey: "",
			Message:    "License key cleared.",
		}
	}

	return domainlicense.SaveResult{
		Success:    true,
		LicenseKey: normalized,
		Message:    "License key saved.",
	}
}
