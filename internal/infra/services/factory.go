package services

import (
	"runtime"

	linuxservices "phant/internal/infra/services/linux"
	"phant/internal/infra/system"
)

// NewProviderForCurrentOS resolves an OS-specific Services provider.
func NewProviderForCurrentOS(runner system.Runner) Provider {
	return NewProviderForOS(runtime.GOOS, runner)
}

// NewProviderForOS resolves an OS-specific Services provider for tests and wiring.
func NewProviderForOS(platform string, runner system.Runner) Provider {
	switch platform {
	case "linux":
		return linuxservices.NewProvider(runner)
	default:
		return newUnsupportedProvider(platform)
	}
}
