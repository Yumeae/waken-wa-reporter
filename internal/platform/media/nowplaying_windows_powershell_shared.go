//go:build windows

package media

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const psTimeout = 6 * time.Second

type psMediaJSON struct {
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	SourceAppID string `json:"sourceAppId"`
}

func getNowPlayingViaPowerShell() (Info, error) {
	const script = `
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$null = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager, Windows.Media.Control, ContentType=WindowsRuntime]
$async = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager]::RequestAsync()
$mgr = $async.GetAwaiter().GetResult()
$session = $mgr.GetCurrentSession()
if (-not $session) { exit 0 }
$info = $session.TryGetMediaPropertiesAsync().GetAwaiter().GetResult()
if (-not $info) { exit 0 }
@{
  title = $info.Title
  artist = $info.Artist
  album = $info.AlbumTitle
  sourceAppId = $session.SourceAppUserModelId
} | ConvertTo-Json -Compress
`
	ctx, cancel := context.WithTimeout(context.Background(), psTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return Info{}, fmt.Errorf("media: powershell timeout: %w", err)
		}
		return Info{}, fmt.Errorf("media: powershell: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	out := strings.TrimSpace(stdout.String())
	if out == "" {
		return Info{}, ErrNoMedia
	}
	var j psMediaJSON
	if err := json.Unmarshal([]byte(out), &j); err != nil {
		return Info{}, fmt.Errorf("media: json: %w", err)
	}
	info := Info{Title: j.Title, Artist: j.Artist, Album: j.Album, SourceAppID: j.SourceAppID}
	if info.IsEmpty() {
		return Info{}, ErrNoMedia
	}
	return info, nil
}
