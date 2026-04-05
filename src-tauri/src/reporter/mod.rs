pub mod activity;
pub mod config;
pub mod foreground;
pub mod media;
pub mod power;
pub mod state;

use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};
use tauri::AppHandle;

use state::AppState;

const DEFAULT_POLL_MS: u64 = 2000;
const DEFAULT_HEARTBEAT_MS: u64 = 60000;

pub async fn run_loop(state: Arc<Mutex<AppState>>, _app: AppHandle) {
    // Load initial config and generate hash key if needed
    {
        let mut s = state.lock().unwrap();
        if s.config.generated_hash_key.is_empty() {
            s.config.generated_hash_key = uuid::Uuid::new_v4().to_string().replace('-', "");
            let _ = config::save_config(&s.config);
        }
        s.status.is_running = true;
    }

    let mut last_snap: Option<(String, String, String)> = None;
    let mut last_report = Instant::now() - Duration::from_secs(3600);

    loop {
        let (poll_ms, heartbeat_ms, cfg) = {
            let s = state.lock().unwrap();
            let poll = s.config.poll_interval_ms.unwrap_or(DEFAULT_POLL_MS as i64) as u64;
            let hb = s
                .config
                .heartbeat_interval_ms
                .unwrap_or(DEFAULT_HEARTBEAT_MS as i64) as u64;
            (poll.max(100), hb, s.config.clone())
        };

        tokio::time::sleep(Duration::from_millis(poll_ms)).await;

        if cfg.api_token.is_empty() || cfg.base_url.is_empty() {
            continue;
        }

        // Gather snapshot
        let fg = foreground::get_snapshot();
        let process_name = fg
            .as_ref()
            .map(|s| s.process_name.clone())
            .unwrap_or_default();
        let process_title = fg
            .as_ref()
            .map(|s| s.process_title.clone())
            .unwrap_or_default();

        let minfo = media::get_now_playing().unwrap_or_default();
        let media_sig = format!("{}|{}|{}", minfo.title, minfo.artist, minfo.album);

        let battery = power::get_battery();

        // Update status
        {
            let mut s = state.lock().unwrap();
            s.status.process_name = process_name.clone();
            s.status.process_title = process_title.clone();
            s.status.media_title = minfo.title.clone();
            s.status.media_artist = minfo.artist.clone();
            s.status.media_album = minfo.album.clone();
            s.status.battery_level = battery.level;
            s.status.is_charging = battery.charging;
        }

        if process_name.is_empty() {
            continue;
        }

        let current_snap = (process_name.clone(), process_title.clone(), media_sig);
        let same = last_snap.as_ref() == Some(&current_snap);

        let should_report = !same
            || (heartbeat_ms > 0
                && last_report.elapsed() >= Duration::from_millis(heartbeat_ms));

        if !should_report {
            continue;
        }

        let req = activity::ReportRequest {
            generated_hash_key: cfg.generated_hash_key.clone(),
            device: if cfg.device_name.is_empty() {
                hostname_or_default()
            } else {
                cfg.device_name.clone()
            },
            device_name: if cfg.device_name.is_empty() {
                hostname_or_default()
            } else {
                cfg.device_name.clone()
            },
            device_type: "desktop".to_string(),
            process_name: process_name.clone(),
            process_title: if process_title.is_empty() {
                None
            } else {
                Some(process_title)
            },
            battery_level: battery.level,
            is_charging: battery.charging,
            push_mode: "realtime".to_string(),
            metadata: build_metadata(&minfo, &cfg),
        };

        let bypass = cfg.bypass_system_proxy;
        let base_url = cfg.base_url.clone();
        let token = cfg.api_token.clone();
        let state_clone = state.clone();

        match activity::post_activity(&base_url, &token, req, bypass).await {
            Ok(_) => {
                last_snap = Some(current_snap);
                last_report = Instant::now();
                let now = chrono::Local::now().format("%H:%M:%S").to_string();
                let mut s = state_clone.lock().unwrap();
                s.status.last_report_time = Some(now);
                s.status.last_error = None;
            }
            Err(e) => {
                let mut s = state_clone.lock().unwrap();
                s.status.last_error = Some(e.to_string());
            }
        }
    }
}

fn hostname_or_default() -> String {
    hostname::get()
        .ok()
        .and_then(|h| h.into_string().ok())
        .unwrap_or_else(|| "desktop".to_string())
}

fn build_metadata(minfo: &media::MediaInfo, _cfg: &config::Config) -> serde_json::Value {
    let mut meta = serde_json::json!({"source": "waken-wa"});
    if !minfo.title.is_empty() || !minfo.artist.is_empty() {
        let mut media_obj = serde_json::Map::new();
        if !minfo.title.is_empty() {
            media_obj.insert(
                "title".to_string(),
                serde_json::Value::String(minfo.title.clone()),
            );
        }
        if !minfo.artist.is_empty() {
            media_obj.insert(
                "artist".to_string(),
                serde_json::Value::String(minfo.artist.clone()),
            );
            media_obj.insert(
                "singer".to_string(),
                serde_json::Value::String(minfo.artist.clone()),
            );
        }
        if !minfo.album.is_empty() {
            media_obj.insert(
                "album".to_string(),
                serde_json::Value::String(minfo.album.clone()),
            );
        }
        meta["media"] = serde_json::Value::Object(media_obj);
        meta["play_source"] = serde_json::Value::String("system_media".to_string());
    }
    meta
}

