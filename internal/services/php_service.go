package services

import (
	"context"
	"time"

	appphpmanager "phant/internal/app/phpmanager"
	domainphpmanager "phant/internal/domain/phpmanager"
	phpinfra "phant/internal/infra/php"
	"phant/internal/infra/system"
)

type PHPService struct {
	service *appphpmanager.Service
}

var phpServiceTimeouts = struct {
	snapshot  time.Duration
	install   time.Duration
	switchV   time.Duration
	settings  time.Duration
	extension time.Duration
}{
	snapshot:  10 * time.Second,
	install:   10 * time.Minute,
	switchV:   30 * time.Second,
	settings:  45 * time.Second,
	extension: 45 * time.Second,
}

func withTimeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration)
}

func NewPHPService() *PHPService {
	provider := phpinfra.NewProviderForCurrentOS(system.NewExecRunner())

	return &PHPService{service: appphpmanager.NewService(appphpmanager.Dependencies{
		Platform:           provider.Platform,
		DiscoverVersions:   provider.DiscoverVersions,
		DiscoverSettings:   provider.DiscoverSettings,
		DiscoverExtensions: provider.DiscoverExtensions,
		InstallVersion:     provider.InstallVersion,
		SwitchVersion:      provider.SwitchVersion,
		UpdateSettings:     provider.UpdateSettings,
		SetExtensionState:  provider.SetExtensionState,
	})}
}

func (s *PHPService) GetPHPManagerSnapshot() domainphpmanager.Snapshot {
	ctx, cancel := withTimeout(phpServiceTimeouts.snapshot)
	defer cancel()

	return s.service.GetSnapshot(ctx)
}

func (s *PHPService) InstallPHPVersion(version string) domainphpmanager.ActionResult {
	ctx, cancel := withTimeout(phpServiceTimeouts.install)
	defer cancel()

	return s.service.InstallVersion(ctx, version)
}

func (s *PHPService) SwitchPHPVersion(version string) domainphpmanager.ActionResult {
	ctx, cancel := withTimeout(phpServiceTimeouts.switchV)
	defer cancel()

	return s.service.SwitchVersion(ctx, version)
}

func (s *PHPService) UpdatePHPIniSettings(request domainphpmanager.IniSettingsUpdateRequest) domainphpmanager.ActionResult {
	ctx, cancel := withTimeout(phpServiceTimeouts.settings)
	defer cancel()

	return s.service.UpdateIniSettings(ctx, request)
}

func (s *PHPService) SetPHPExtensionState(request domainphpmanager.ExtensionToggleRequest) domainphpmanager.ActionResult {
	ctx, cancel := withTimeout(phpServiceTimeouts.extension)
	defer cancel()

	return s.service.SetExtensionState(ctx, request)
}
