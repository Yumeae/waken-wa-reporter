#[derive(Debug, Default, Clone)]
pub struct BatteryInfo {
    pub level: Option<i64>,
    pub charging: Option<bool>,
}

pub fn get_battery() -> BatteryInfo {
    #[cfg(target_os = "windows")]
    {
        get_battery_windows()
    }
    #[cfg(target_os = "macos")]
    {
        get_battery_macos()
    }
    #[cfg(not(any(target_os = "windows", target_os = "macos")))]
    {
        BatteryInfo::default()
    }
}

#[cfg(target_os = "windows")]
fn get_battery_windows() -> BatteryInfo {
    use windows::Win32::System::Power::{GetSystemPowerStatus, SYSTEM_POWER_STATUS};
    unsafe {
        let mut status = SYSTEM_POWER_STATUS::default();
        if GetSystemPowerStatus(&mut status).is_ok() {
            let level = if status.BatteryLifePercent <= 100 {
                Some(status.BatteryLifePercent as i64)
            } else {
                None
            };
            let charging = match status.ACLineStatus {
                0 => Some(false),
                1 => Some(true),
                _ => None,
            };
            BatteryInfo { level, charging }
        } else {
            BatteryInfo::default()
        }
    }
}

#[cfg(target_os = "macos")]
fn get_battery_macos() -> BatteryInfo {
    use std::process::Command;
    let output = Command::new("pmset").args(["-g", "batt"]).output();
    if let Ok(out) = output {
        if out.status.success() {
            let s = String::from_utf8_lossy(&out.stdout);
            let charging =
                s.contains("charging") && !s.contains("discharging") && !s.contains("not charging");
            let level = s
                .lines()
                .find(|l| l.contains('%'))
                .and_then(|l| l.split('%').next())
                .and_then(|l| l.split_whitespace().last())
                .and_then(|s| s.trim_end_matches('%').parse::<i64>().ok());
            return BatteryInfo {
                level,
                charging: Some(charging),
            };
        }
    }
    BatteryInfo::default()
}
