package php

import (
	"runtime"

	linuxphp "phant/internal/infra/php/linux"
	"phant/internal/infra/system"
)

// NewProviderForCurrentOS resolves an OS-specific PHP provider.
func NewProviderForCurrentOS(runner system.Runner) Provider {
	return NewProviderForOS(runtime.GOOS, runner)
}

// NewProviderForOS resolves an OS-specific PHP provider for tests and wiring.
func NewProviderForOS(platform string, runner system.Runner) Provider {
	switch platform {
	case "linux":
		return linuxphp.NewProvider(runner)
	default:
		return newUnsupportedProvider(platform)
	}
}
