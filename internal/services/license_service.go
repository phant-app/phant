package services

import (
	"context"
	"time"

	applicense "phant/internal/app/license"
	domainlicense "phant/internal/domain/license"
	settingsinfra "phant/internal/infra/settings"
)

const licenseServiceTimeout = 5 * time.Second

type LicenseService struct {
	service *applicense.Service
}

func NewLicenseService() *LicenseService {
	provider := settingsinfra.NewFileProvider()
	return &LicenseService{
		service: applicense.NewService(applicense.Dependencies{
			LoadSettings: provider.Load,
			SaveSettings: provider.Save,
		}),
	}
}

func (s *LicenseService) GetLicenseKey() domainlicense.KeyResult {
	ctx, cancel := context.WithTimeout(context.Background(), licenseServiceTimeout)
	defer cancel()
	return s.service.GetKey(ctx)
}

func (s *LicenseService) SaveLicenseKey(licenseKey string) domainlicense.SaveResult {
	ctx, cancel := context.WithTimeout(context.Background(), licenseServiceTimeout)
	defer cancel()
	return s.service.SaveKey(ctx, licenseKey)
}
