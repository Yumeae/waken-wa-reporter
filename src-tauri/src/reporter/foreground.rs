#[derive(Debug, Default, Clone)]
pub struct ForegroundSnapshot {
    pub process_name: String,
    pub process_title: String,
}

pub fn get_snapshot() -> Option<ForegroundSnapshot> {
    #[cfg(target_os = "windows")]
    {
        get_snapshot_windows()
    }
    #[cfg(target_os = "macos")]
    {
        get_snapshot_macos()
    }
    #[cfg(not(any(target_os = "windows", target_os = "macos")))]
    {
        None
    }
}

#[cfg(target_os = "windows")]
fn get_snapshot_windows() -> Option<ForegroundSnapshot> {
    use windows::Win32::UI::WindowsAndMessaging::{
        GetForegroundWindow, GetWindowTextW, GetWindowThreadProcessId,
    };
    use windows::Win32::System::Threading::{
        OpenProcess, QueryFullProcessImageNameW, PROCESS_NAME_WIN32,
        PROCESS_QUERY_LIMITED_INFORMATION,
    };
    use windows::Win32::Foundation::CloseHandle;
    use windows::core::PWSTR;

    unsafe {
        let hwnd = GetForegroundWindow();
        if hwnd.0 as usize == 0 {
            return None;
        }

        // Window title
        let mut title_buf = vec![0u16; 512];
        let title_len = GetWindowTextW(hwnd, &mut title_buf) as usize;
        let process_title = String::from_utf16_lossy(&title_buf[..title_len]);

        // PID
        let mut pid: u32 = 0;
        GetWindowThreadProcessId(hwnd, Some(&mut pid));
        if pid == 0 {
            return None;
        }

        // Process name
        let handle = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, pid).ok()?;
        let mut name_buf = vec![0u16; 512];
        let mut size = name_buf.len() as u32;
        let process_name = if QueryFullProcessImageNameW(
            handle,
            PROCESS_NAME_WIN32,
            PWSTR(name_buf.as_mut_ptr()),
            &mut size,
        )
        .is_ok()
        {
            let path = String::from_utf16_lossy(&name_buf[..size as usize]);
            std::path::Path::new(&path)
                .file_name()
                .and_then(|n| n.to_str())
                .unwrap_or(&path)
                .to_string()
        } else {
            String::new()
        };
        let _ = CloseHandle(handle);

        if process_name.is_empty() {
            return None;
        }

        Some(ForegroundSnapshot {
            process_name,
            process_title,
        })
    }
}

#[cfg(target_os = "macos")]
fn get_snapshot_macos() -> Option<ForegroundSnapshot> {
    use std::process::Command;
    let output = Command::new("osascript")
        .arg("-e")
        .arg("tell application \"System Events\" to get name of first application process whose frontmost is true")
        .output()
        .ok()?;
    if output.status.success() {
        let name = String::from_utf8_lossy(&output.stdout).trim().to_string();
        if !name.is_empty() {
            return Some(ForegroundSnapshot {
                process_name: name,
                process_title: String::new(),
            });
        }
    }
    None
}
