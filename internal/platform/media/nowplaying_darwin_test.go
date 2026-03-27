//go:build darwin

package media

import (
	"errors"
	"testing"
)

func TestGetNowPlaying_Darwin(t *testing.T) {
	info, err := GetNowPlaying()
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			t.Skip("CGO disabled; now playing unavailable")
		}
		if errors.Is(err, ErrNoMedia) {
			t.Skip("no MediaRemote now playing info")
		}
		t.Skipf("GetNowPlaying: %v", err)
	}
	if info.IsEmpty() {
		t.Fatal("non-error result must be non-empty")
	}
	t.Logf("media title=%q artist=%q album=%q", info.Title, info.Artist, info.Album)
}
