package settings

import "runtime"

func runtimePlatform() string {
	return runtime.GOOS
}
