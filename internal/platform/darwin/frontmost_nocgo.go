//go:build darwin && !cgo

package darwin

// GetForegroundApplicationName returns failure when CGO is disabled (e.g. IDE analysis on non-macOS or cross-compile without cgo).
func GetForegroundApplicationName() (string, bool) {
	return "", true
}
