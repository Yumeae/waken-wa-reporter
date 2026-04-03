package media

import (
	"errors"
	"strings"
)

// ErrUnsupported is returned on platforms without a now-playing implementation.
var ErrUnsupported = errors.New("media: now playing not supported on this platform")

// ErrNoMedia is returned when the OS reports no active media session or empty metadata.
var ErrNoMedia = errors.New("media: no active media metadata")

// Info holds normalized now-playing fields for activity metadata.
type Info struct {
	Title       string
	Artist      string
	Album       string
	SourceAppID string
}

// IsEmpty reports whether all fields are empty after trimming.
func (i Info) IsEmpty() bool {
	return strings.TrimSpace(i.Title) == "" &&
		strings.TrimSpace(i.Artist) == "" &&
		strings.TrimSpace(i.Album) == ""
}

// Signature returns a stable string for change detection (foreground unchanged but track changed).
func (i Info) Signature() string {
	if i.IsEmpty() {
		return ""
	}
	return strings.TrimSpace(i.Title) + "\x1e" + strings.TrimSpace(i.Artist) + "\x1e" + strings.TrimSpace(i.Album) + "\x1e" + strings.TrimSpace(i.SourceAppID)
}

// AsMap returns metadata.media-compatible keys: title, artist, album, and singer (alias of artist).
func (i Info) AsMap() map[string]any {
	if i.IsEmpty() {
		return nil
	}
	t := strings.TrimSpace(i.Title)
	a := strings.TrimSpace(i.Artist)
	al := strings.TrimSpace(i.Album)
	m := make(map[string]any, 4)
	if t != "" {
		m["title"] = t
	}
	if a != "" {
		m["artist"] = a
		m["singer"] = a
	}
	if al != "" {
		m["album"] = al
	}
	return m
}
