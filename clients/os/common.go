package os

import "runtime"

// isARM returns true if running on ARM architecture
func isARM() bool {
	return runtime.GOARCH == "arm" || runtime.GOARCH == "arm64"
}
