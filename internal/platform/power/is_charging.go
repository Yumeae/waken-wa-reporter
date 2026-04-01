//go:build !windows && !darwin

package power

// IsCharging returns whether the system is charging (AC power), or nil if unknown.
//
// This is a stub implementation for unsupported platforms; see OS-specific files.
func IsCharging() *bool { return nil }
