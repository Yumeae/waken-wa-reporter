//go:build !windows && !darwin

package media

import (
	"errors"
	"testing"
)

func TestGetNowPlaying_Stub(t *testing.T) {
	_, err := GetNowPlaying()
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("got %v want ErrUnsupported", err)
	}
}
