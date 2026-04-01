//go:build darwin && !cgo

package power

func IsCharging() *bool { return nil }
