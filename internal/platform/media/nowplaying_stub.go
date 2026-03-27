//go:build !windows && !darwin

package media

// GetNowPlaying returns ErrUnsupported on non-desktop or unsupported OS targets.
func GetNowPlaying() (Info, error) {
	return Info{}, ErrUnsupported
}
