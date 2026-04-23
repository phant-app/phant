package services

import (
	"context"
	"strings"
	"time"

	applicense "phant/internal/app/license"
	appupdate "phant/internal/app/update"
	domainupdate "phant/internal/domain/update"
	settingsinfra "phant/internal/infra/settings"
	"phant/internal/infra/system"
	updateinfra "phant/internal/infra/update"
)

var BuildVersion = "0.0.1"

const updateServiceTimeout = 60 * time.Second

type UpdateService struct {
	service *appupdate.Service
}

func NewUpdateService() *UpdateService {
	settingsProvider := settingsinfra.NewFileProvider()
	licenseService := applicense.NewService(applicense.Dependencies{
		LoadSettings: settingsProvider.Load,
		SaveSettings: settingsProvider.Save,
	})
	updateProvider := updateinfra.NewProviderForCurrentOS(system.NewExecRunner())

	return &UpdateService{service: appupdate.NewService(appupdate.Dependencies{
		CurrentVersion:    func() string { return BuildVersion },
		Platform:          updateProvider.Platform,
		GetLicenseKey:     licenseService.GetKey,
		HTTPClient:        updateProvider.HTTPClient,
		InstallDownloaded: updateProvider.InstallDownloaded,
	})}
}

func (s *UpdateService) CurrentVersion() string {
	return normalizeVersion(BuildVersion)
}

func (s *UpdateService) CheckForUpdate(manifestURL string) domainupdate.CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), updateServiceTimeout)
	defer cancel()
	return s.service.CheckForUpdate(ctx, manifestURL)
}

func (s *UpdateService) DownloadLatest(manifestURL string) domainupdate.DownloadResult {
	ctx, cancel := context.WithTimeout(context.Background(), updateServiceTimeout)
	defer cancel()
	return s.service.DownloadLatest(ctx, manifestURL)
}

func (s *UpdateService) InstallDownloaded(downloadedPath string) domainupdate.InstallResult {
	ctx, cancel := context.WithTimeout(context.Background(), updateServiceTimeout)
	defer cancel()
	return s.service.InstallDownloaded(ctx, downloadedPath)
}

func normalizeVersion(v string) string {
	normalized := strings.TrimSpace(v)
	normalized = strings.TrimPrefix(normalized, "v")
	if idx := strings.IndexAny(normalized, "-+"); idx >= 0 {
		normalized = normalized[:idx]
	}
	return normalized
}
