// Package openurl starts the system default browser or handler for an https URL.
package openurl

import (
	"fmt"
	"os/exec"
	"runtime"
)

// InBrowser opens url in the default browser (best-effort).
func InBrowser(url string) error {
	if url == "" {
		return fmt.Errorf("openurl: empty url")
	}
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
