package update

import "runtime"

func runtimePlatform() string {
	return runtime.GOOS
}
