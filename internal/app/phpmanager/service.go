package phpmanager

import (
	"context"
	"time"

	domainphpmanager "phant/internal/domain/phpmanager"
)

type Dependencies struct {
	Now                func() time.Time
	Platform           func() string
	DiscoverVersions   func(context.Context) (string, []domainphpmanager.Version, error)
	DiscoverSettings   func(context.Context) (domainphpmanager.IniSettings, error)
	DiscoverExtensions func(context.Context) ([]domainphpmanager.Extension, error)
	InstallVersion     func(context.Context, string) domainphpmanager.ActionResult
	SwitchVersion      func(context.Context, string) domainphpmanager.ActionResult
	UpdateSettings     func(context.Context, domainphpmanager.IniSettingsUpdateRequest) domainphpmanager.ActionResult
	SetExtensionState  func(context.Context, domainphpmanager.ExtensionToggleRequest) domainphpmanager.ActionResult
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

func (s *Service) GetSnapshot(ctx context.Context) domainphpmanager.Snapshot {
	report := domainphpmanager.Snapshot{
		GeneratedAt: s.deps.Now().UTC().Format(time.RFC3339),
		Platform:    s.deps.Platform(),
		Supported:   s.deps.Platform() == "linux",
	}

	if !report.Supported {
		report.Warnings = append(report.Warnings, "PHP manager is currently supported on Linux only.")
		return report
	}

	if s.deps.DiscoverVersions != nil {
		activeVersion, versions, err := s.deps.DiscoverVersions(ctx)
		report.ActiveVersion = activeVersion
		report.Versions = versions
		if err != nil {
			report.LastError = err.Error()
		}
	}

	if s.deps.DiscoverSettings != nil {
		settings, err := s.deps.DiscoverSettings(ctx)
		report.Settings = settings
		if err != nil {
			report.Warnings = append(report.Warnings, err.Error())
		}
	}

	if s.deps.DiscoverExtensions != nil {
		extensions, err := s.deps.DiscoverExtensions(ctx)
		report.Extensions = extensions
		if err != nil {
			report.Warnings = append(report.Warnings, err.Error())
		}
	}

	return report
}

func (s *Service) InstallVersion(ctx context.Context, version string) domainphpmanager.ActionResult {
	if s.deps.InstallVersion == nil {
		return domainphpmanager.ActionResult{Supported: false, Message: "install operation is unavailable"}
	}
	return s.deps.InstallVersion(ctx, version)
}

func (s *Service) SwitchVersion(ctx context.Context, version string) domainphpmanager.ActionResult {
	if s.deps.SwitchVersion == nil {
		return domainphpmanager.ActionResult{Supported: false, Message: "switch operation is unavailable"}
	}
	return s.deps.SwitchVersion(ctx, version)
}

func (s *Service) UpdateIniSettings(ctx context.Context, request domainphpmanager.IniSettingsUpdateRequest) domainphpmanager.ActionResult {
	if s.deps.UpdateSettings == nil {
		return domainphpmanager.ActionResult{Supported: false, Message: "settings update operation is unavailable"}
	}
	return s.deps.UpdateSettings(ctx, request)
}

func (s *Service) SetExtensionState(ctx context.Context, request domainphpmanager.ExtensionToggleRequest) domainphpmanager.ActionResult {
	if s.deps.SetExtensionState == nil {
		return domainphpmanager.ActionResult{Supported: false, Message: "extension operation is unavailable"}
	}
	return s.deps.SetExtensionState(ctx, request)
}
