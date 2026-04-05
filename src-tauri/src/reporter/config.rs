use serde::{Deserialize, Serialize};
use std::path::PathBuf;

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Config {
    #[serde(default)]
    pub base_url: String,
    #[serde(default)]
    pub api_token: String,
    #[serde(default)]
    pub device_name: String,
    #[serde(default)]
    pub generated_hash_key: String,
    #[serde(default)]
    pub poll_interval_ms: Option<i64>,
    #[serde(default)]
    pub heartbeat_interval_ms: Option<i64>,
    #[serde(default)]
    pub bypass_system_proxy: bool,
}

pub fn config_path() -> Option<PathBuf> {
    dirs::config_dir().map(|d| d.join("waken-wa").join("config.json"))
}

pub fn load_config() -> Option<Config> {
    let path = config_path()?;
    let data = std::fs::read_to_string(&path).ok()?;
    serde_json::from_str(&data).ok()
}

pub fn save_config(cfg: &Config) -> Result<(), Box<dyn std::error::Error>> {
    let path = config_path().ok_or("Could not determine config path")?;
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }
    let data = serde_json::to_string_pretty(cfg)?;
    std::fs::write(&path, data)?;
    Ok(())
}
