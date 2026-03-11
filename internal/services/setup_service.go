package services

import (
	"context"

	"phant/internal/setup"
)

type SetupService struct {
	runtime *collectorRuntime
}

func (s *SetupService) GetSetupDiagnostics() setup.Diagnostics {
	return setup.CollectDiagnostics(context.Background())
}

func (s *SetupService) EnableCLIHook() setup.HookInstallResult {
	return setup.InstallCLIHook(context.Background(), s.runtime.collectorSocketPath())
}

func (s *SetupService) GetValetLinuxVerification() setup.ValetLinuxVerification {
	return setup.VerifyValetLinux(context.Background())
}

func (s *SetupService) GetValetSites() setup.ValetSitesResult {
	return setup.DiscoverValetSites(context.Background())
}

func (s *SetupService) ApplyValetLinuxRemediation(confirm bool) setup.ValetLinuxRemediationResult {
	return setup.ApplyValetLinuxRemediation(context.Background(), confirm)
}
