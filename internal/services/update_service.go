package services

import (
	"context"
	"runtime"
	"strings"
	"time"

	applicense "phant/internal/app/license"
	appupdate "phant/internal/app/update"
	domainupdate "phant/internal/domain/update"
	settingsinfra "phant/internal/infra/settings"
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
	updateProvider := updateinfra.NewProviderForCurrentOS()

	return &UpdateService{service: appupdate.NewService(appupdate.Dependencies{
		CurrentVersion: func() string { return BuildVersion },
		Platform:       func() string { return runtime.GOOS },
		GetLicenseKey:  licenseService.GetKey,
		HTTPClient:     updateProvider.HTTPClient,
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

func normalizeVersion(v string) string {
	normalized := strings.TrimSpace(v)
	normalized = strings.TrimPrefix(normalized, "v")
	if idx := strings.IndexAny(normalized, "-+"); idx >= 0 {
		normalized = normalized[:idx]
	}
	return normalized
}
