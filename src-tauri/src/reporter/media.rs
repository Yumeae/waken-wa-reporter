#[derive(Debug, Default, Clone)]
pub struct MediaInfo {
    pub title: String,
    pub artist: String,
    pub album: String,
}

pub fn get_now_playing() -> Option<MediaInfo> {
    #[cfg(target_os = "windows")]
    {
        get_now_playing_windows()
    }
    #[cfg(target_os = "macos")]
    {
        get_now_playing_macos()
    }
    #[cfg(not(any(target_os = "windows", target_os = "macos")))]
    {
        None
    }
}

#[cfg(target_os = "windows")]
fn get_now_playing_windows() -> Option<MediaInfo> {
    use std::process::Command;
    let script = r#"
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
$mgr = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager,Windows.Media.Control,ContentType=WindowsRuntime]
$task = $mgr::RequestAsync()
$task.AsTask().Wait(5000) | Out-Null
if ($task.Status -ne 'RanToCompletion') { exit 1 }
$sessions = $task.Result.GetSessions()
if ($sessions.Count -eq 0) { exit 1 }
$s = $sessions[0]
$infoTask = $s.TryGetMediaPropertiesAsync()
$infoTask.AsTask().Wait(5000) | Out-Null
if ($infoTask.Status -ne 'RanToCompletion') { exit 1 }
$info = $infoTask.Result
[PSCustomObject]@{title=$info.Title;artist=$info.Artist;album=$info.AlbumTitle} | ConvertTo-Json
"#;
    let output = Command::new("powershell")
        .args(["-NoProfile", "-NonInteractive", "-Command", script])
        .output()
        .ok()?;
    if !output.status.success() {
        return None;
    }
    let json: serde_json::Value = serde_json::from_slice(&output.stdout).ok()?;
    let info = MediaInfo {
        title: json["title"].as_str().unwrap_or("").trim().to_string(),
        artist: json["artist"].as_str().unwrap_or("").trim().to_string(),
        album: json["album"].as_str().unwrap_or("").trim().to_string(),
    };
    if info.title.is_empty() && info.artist.is_empty() {
        return None;
    }
    Some(info)
}

#[cfg(target_os = "macos")]
fn get_now_playing_macos() -> Option<MediaInfo> {
    use std::process::Command;
    let script = r#"
tell application "System Events"
    set nowPlayingApps to every process whose name is "Music" or name is "Spotify" or name is "Vox"
    if (count nowPlayingApps) > 0 then
        tell application "Music"
            if player state is playing then
                return (name of current track) & "|||" & (artist of current track) & "|||" & (album of current track)
            end if
        end tell
    end if
end tell
return ""
"#;
    let output = Command::new("osascript")
        .arg("-e")
        .arg(script)
        .output()
        .ok()?;
    if output.status.success() {
        let s = String::from_utf8_lossy(&output.stdout).trim().to_string();
        if !s.is_empty() {
            let parts: Vec<&str> = s.split("|||").collect();
            return Some(MediaInfo {
                title: parts.first().copied().unwrap_or("").to_string(),
                artist: parts.get(1).copied().unwrap_or("").to_string(),
                album: parts.get(2).copied().unwrap_or("").to_string(),
            });
        }
    }
    None
}
