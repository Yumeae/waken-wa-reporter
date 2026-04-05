import { useState, useEffect, useCallback } from "react";
import { invoke } from "@tauri-apps/api/core";
import Dashboard from "./components/Dashboard";
import Settings from "./components/Settings";

export interface Config {
  base_url: string;
  api_token: string;
  device_name: string;
  generated_hash_key: string;
  poll_interval_ms: number | null;
  heartbeat_interval_ms: number | null;
  bypass_system_proxy: boolean;
}

export interface ActivityStatus {
  process_name: string;
  process_title: string;
  media_title: string;
  media_artist: string;
  media_album: string;
  battery_level: number | null;
  is_charging: boolean | null;
  is_running: boolean;
  last_report_time: string | null;
  last_error: string | null;
}

type Page = "dashboard" | "settings";

export default function App() {
  const [page, setPage] = useState<Page>("dashboard");
  const [config, setConfig] = useState<Config | null>(null);
  const [status, setStatus] = useState<ActivityStatus | null>(null);
  const [saving, setSaving] = useState(false);
  const [saveMsg, setSaveMsg] = useState<string | null>(null);

  const loadConfig = useCallback(async () => {
    try {
      const cfg = await invoke<Config>("get_config");
      setConfig(cfg);
    } catch (e) {
      console.error("Failed to load config:", e);
    }
  }, []);

  const loadStatus = useCallback(async () => {
    try {
      const s = await invoke<ActivityStatus>("get_status");
      setStatus(s);
    } catch (e) {
      console.error("Failed to get status:", e);
    }
  }, []);

  useEffect(() => {
    loadConfig();
    loadStatus();
    const interval = setInterval(loadStatus, 2000);
    return () => clearInterval(interval);
  }, [loadConfig, loadStatus]);

  const saveConfig = async (cfg: Config) => {
    setSaving(true);
    setSaveMsg(null);
    try {
      await invoke("save_config", { config: cfg });
      setConfig(cfg);
      setSaveMsg("配置已保存");
      setTimeout(() => setSaveMsg(null), 3000);
    } catch (e) {
      setSaveMsg(`保存失败：${e}`);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col h-screen bg-slate-900 overflow-hidden">
      {/* Title bar area */}
      <div
        data-tauri-drag-region
        className="flex items-center justify-between px-4 py-3 bg-slate-900 border-b border-slate-700/50 select-none"
      >
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-sky-500 shadow-lg shadow-sky-500/30" />
          <span className="text-sm font-semibold text-white tracking-wide">
            Waken-Wa Reporter
          </span>
        </div>
        <nav className="flex gap-1">
          <button
            onClick={() => setPage("dashboard")}
            className={`px-3 py-1 text-xs rounded-md transition-colors ${
              page === "dashboard"
                ? "bg-sky-600 text-white"
                : "text-slate-400 hover:text-white hover:bg-slate-700"
            }`}
          >
            状态
          </button>
          <button
            onClick={() => setPage("settings")}
            className={`px-3 py-1 text-xs rounded-md transition-colors ${
              page === "settings"
                ? "bg-sky-600 text-white"
                : "text-slate-400 hover:text-white hover:bg-slate-700"
            }`}
          >
            设置
          </button>
        </nav>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto scrollbar-thin">
        {page === "dashboard" && (
          <Dashboard status={status} config={config} />
        )}
        {page === "settings" && config && (
          <Settings
            config={config}
            onSave={saveConfig}
            saving={saving}
            saveMsg={saveMsg}
          />
        )}
      </div>
    </div>
  );
}
