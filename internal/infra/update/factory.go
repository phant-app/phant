package update

import (
	"runtime"

	"phant/internal/infra/system"
	linuxupdate "phant/internal/infra/update/linux"
)

// NewProviderForCurrentOS resolves an OS-specific update provider.
func NewProviderForCurrentOS(runner system.Runner) Provider {
	return NewProviderForOS(runtime.GOOS, runner)
}

// NewProviderForOS resolves an OS-specific update provider for tests and wiring.
func NewProviderForOS(platform string, runner system.Runner) Provider {
	switch platform {
	case "linux":
		return linuxupdate.NewProvider(runner)
	default:
		return newUnsupportedProvider(platform)
	}
}
