use serde::{Deserialize, Serialize};
use super::config::Config;

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ActivityStatus {
    pub process_name: String,
    pub process_title: String,
    pub media_title: String,
    pub media_artist: String,
    pub media_album: String,
    pub battery_level: Option<i64>,
    pub is_charging: Option<bool>,
    pub is_running: bool,
    pub last_report_time: Option<String>,
    pub last_error: Option<String>,
}

pub struct AppState {
    pub config: Config,
    pub status: ActivityStatus,
}

impl AppState {
    pub fn new() -> Self {
        let config = super::config::load_config().unwrap_or_default();
        Self {
            config,
            status: ActivityStatus::default(),
        }
    }
}
