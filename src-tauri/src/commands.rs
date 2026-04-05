use std::sync::{Arc, Mutex};
use tauri::State;

use crate::reporter::{config::Config, state::AppState};

#[tauri::command]
pub async fn get_config(state: State<'_, Arc<Mutex<AppState>>>) -> Result<Config, String> {
    let s = state.lock().map_err(|e| e.to_string())?;
    Ok(s.config.clone())
}

#[tauri::command]
pub async fn save_config(
    state: State<'_, Arc<Mutex<AppState>>>,
    config: Config,
) -> Result<(), String> {
    let mut s = state.lock().map_err(|e| e.to_string())?;
    s.config = config.clone();
    crate::reporter::config::save_config(&config).map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_status(state: State<'_, Arc<Mutex<AppState>>>) -> Result<serde_json::Value, String> {
    let s = state.lock().map_err(|e| e.to_string())?;
    serde_json::to_value(&s.status).map_err(|e| e.to_string())
}
