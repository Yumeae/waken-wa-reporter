//go:build darwin && !cgo

package media

// GetNowPlaying returns ErrUnsupported when CGO is disabled (e.g. some cross-compile toolchains).
func GetNowPlaying() (Info, error) {
	return Info{}, ErrUnsupported
}
