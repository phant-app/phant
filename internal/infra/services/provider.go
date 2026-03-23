package services

import (
	"context"
	"fmt"

	"phant/internal/domain/servicesstatus"
)

// Provider defines OS-specific services status discovery.
type Provider interface {
	Platform() string
	DiscoverServices(ctx context.Context) ([]servicesstatus.ServiceStatus, []string, error)
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

func (p unsupportedProvider) DiscoverServices(context.Context) ([]servicesstatus.ServiceStatus, []string, error) {
	warnings := []string{fmt.Sprintf("Services status is currently unsupported on %s.", p.platform)}
	return []servicesstatus.ServiceStatus{}, warnings, nil
}
