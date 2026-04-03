//go:build windows && waken_powershell_media

package media

// GetNowPlaying uses the PowerShell GSMTC reader directly when the build tag is enabled.
func GetNowPlaying() (Info, error) {
	return getNowPlayingViaPowerShell()
}
