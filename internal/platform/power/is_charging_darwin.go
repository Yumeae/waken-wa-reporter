//go:build darwin && cgo

package power

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/ps/IOPowerSources.h>
#include <IOKit/ps/IOPSKeys.h>

static int waken_is_charging(int *known, int *charging) {
  *known = 0;
  *charging = 0;

  CFTypeRef info = IOPSCopyPowerSourcesInfo();
  if (info == NULL) return 0;

  CFArrayRef list = IOPSCopyPowerSourcesList(info);
  if (list == NULL) {
    CFRelease(info);
    return 0;
  }

  CFIndex n = CFArrayGetCount(list);
  for (CFIndex i = 0; i < n; i++) {
    CFTypeRef ps = CFArrayGetValueAtIndex(list, i);
    if (ps == NULL) continue;

    CFDictionaryRef desc = IOPSGetPowerSourceDescription(info, ps);
    if (desc == NULL) continue;

    // Prefer kIOPSIsChargingKey when present.
    CFBooleanRef isCharging = (CFBooleanRef)CFDictionaryGetValue(desc, CFSTR(kIOPSIsChargingKey));
    if (isCharging != NULL) {
      *known = 1;
      *charging = CFBooleanGetValue(isCharging) ? 1 : 0;
      break;
    }

    // Fallback: infer from kIOPSPowerSourceStateKey == kIOPSACPowerValue.
    CFStringRef state = (CFStringRef)CFDictionaryGetValue(desc, CFSTR(kIOPSPowerSourceStateKey));
    if (state != NULL) {
      if (CFStringCompare(state, CFSTR(kIOPSACPowerValue), 0) == kCFCompareEqualTo) {
        *known = 1;
        *charging = 1;
        break;
      }
      if (CFStringCompare(state, CFSTR(kIOPSBatteryPowerValue), 0) == kCFCompareEqualTo) {
        *known = 1;
        *charging = 0;
        break;
      }
    }
  }

  CFRelease(list);
  CFRelease(info);
  return 1;
}
*/
import "C"

func IsCharging() *bool {
	var known C.int
	var charging C.int
	ok := C.waken_is_charging(&known, &charging)
	if ok == 0 || known == 0 {
		return nil
	}
	v := charging != 0
	return &v
}
