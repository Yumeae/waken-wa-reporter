//go:build windows

package power

import (
	"syscall"
	"unsafe"
)

// https://learn.microsoft.com/windows/win32/api/winbase/ns-winbase-system_power_status
type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

var (
	modkernel32              = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemPowerStatus = modkernel32.NewProc("GetSystemPowerStatus")
)

func IsCharging() *bool {
	var s systemPowerStatus
	r1, _, _ := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&s)))
	if r1 == 0 {
		return nil
	}
	switch s.ACLineStatus {
	case 0:
		v := false
		return &v
	case 1:
		v := true
		return &v
	default:
		return nil
	}
}
