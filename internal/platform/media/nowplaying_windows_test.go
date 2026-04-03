//go:build windows && !waken_powershell_media

package media

import (
	"errors"
	"testing"
)

func TestGetNowPlaying_Windows(t *testing.T) {
	info, err := GetNowPlaying()
	if err != nil {
		if errors.Is(err, ErrNoMedia) {
			t.Skip("no GSMTC session or empty metadata")
		}
		t.Skipf("GetNowPlaying: %v", err)
	}
	if info.IsEmpty() {
		t.Fatal("non-error result must be non-empty")
	}
	t.Logf("media title=%q artist=%q album=%q sourceAppID=%q", info.Title, info.Artist, info.Album, info.SourceAppID)
}
