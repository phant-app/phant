package php

import (
	"context"
	"fmt"

	domainphpmanager "phant/internal/domain/phpmanager"
)

// Provider defines OS-specific PHP manager operations.
type Provider interface {
	Platform() string
	DiscoverVersions(ctx context.Context) (string, []domainphpmanager.Version, error)
	DiscoverSettings(ctx context.Context) (domainphpmanager.IniSettings, error)
	DiscoverExtensions(ctx context.Context) ([]domainphpmanager.Extension, error)
	InstallVersion(ctx context.Context, version string) domainphpmanager.ActionResult
	SwitchVersion(ctx context.Context, version string) domainphpmanager.ActionResult
	UpdateSettings(ctx context.Context, request domainphpmanager.IniSettingsUpdateRequest) domainphpmanager.ActionResult
	SetExtensionState(ctx context.Context, request domainphpmanager.ExtensionToggleRequest) domainphpmanager.ActionResult
}

type unsupportedProvider struct {
	platform string
}

func newUnsupportedProvider(platform string) Provider {
	return unsupportedProvider{platform: platform}
}

func (p unsupportedProvider) Platform() string {
	return p.platform
}

func (p unsupportedProvider) DiscoverVersions(context.Context) (string, []domainphpmanager.Version, error) {
	return "", nil, nil
}

func (p unsupportedProvider) DiscoverSettings(context.Context) (domainphpmanager.IniSettings, error) {
	return domainphpmanager.IniSettings{}, nil
}

func (p unsupportedProvider) DiscoverExtensions(context.Context) ([]domainphpmanager.Extension, error) {
	return nil, nil
}

func (p unsupportedProvider) InstallVersion(_ context.Context, version string) domainphpmanager.ActionResult {
	return unsupportedActionResult(p.platform, "install", version)
}

func (p unsupportedProvider) SwitchVersion(_ context.Context, version string) domainphpmanager.ActionResult {
	return unsupportedActionResult(p.platform, "switch", version)
}

func (p unsupportedProvider) UpdateSettings(_ context.Context, _ domainphpmanager.IniSettingsUpdateRequest) domainphpmanager.ActionResult {
	return unsupportedActionResult(p.platform, "update-settings", "")
}

func (p unsupportedProvider) SetExtensionState(_ context.Context, request domainphpmanager.ExtensionToggleRequest) domainphpmanager.ActionResult {
	return unsupportedActionResult(p.platform, "toggle-extension", request.Name)
}

func unsupportedActionResult(platform string, action string, version string) domainphpmanager.ActionResult {
	return domainphpmanager.ActionResult{
		Supported: false,
		Version:   version,
		Message:   fmt.Sprintf("PHP manager action %q is currently unsupported on %s.", action, platform),
	}
}
