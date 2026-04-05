import { ActivityStatus, Config } from "../App";

interface Props {
  status: ActivityStatus | null;
  config: Config | null;
}

function StatusDot({ active }: { active: boolean }) {
  return (
    <span
      className={`inline-block w-2 h-2 rounded-full ${
        active ? "bg-green-400 shadow-md shadow-green-400/50 animate-pulse" : "bg-slate-500"
      }`}
    />
  );
}

function Card({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-slate-800/60 border border-slate-700/50 rounded-xl p-4 backdrop-blur-sm">
      <div className="text-xs font-medium text-slate-400 uppercase tracking-wider mb-3">
        {title}
      </div>
      {children}
    </div>
  );
}

function BatteryIcon({ level, charging }: { level: number | null; charging: boolean | null }) {
  if (level === null) return <span className="text-slate-500 text-sm">未知</span>;

  const color =
    charging === true ? "text-green-400" :
    level > 30 ? "text-emerald-400" :
    level > 15 ? "text-yellow-400" : "text-red-400";

  return (
    <span className={`font-semibold ${color}`}>
      {level}%{charging === true ? " ⚡" : ""}
    </span>
  );
}

export default function Dashboard({ status, config }: Props) {
  if (!status) {
    return (
      <div className="flex items-center justify-center h-full text-slate-500 text-sm">
        正在加载...
      </div>
    );
  }

  const hasMedia = status.media_title || status.media_artist;

  return (
    <div className="p-4 space-y-3">
      {/* Reporter Status */}
      <Card title="上报状态">
        <div className="flex items-center gap-3">
          <StatusDot active={status.is_running} />
          <div>
            <div className="text-sm font-medium text-white">
              {status.is_running ? "运行中" : "已停止"}
            </div>
            {status.last_report_time && (
              <div className="text-xs text-slate-400 mt-0.5">
                上次上报：{status.last_report_time}
              </div>
            )}
            {status.last_error && (
              <div className="text-xs text-red-400 mt-1 truncate max-w-[280px]" title={status.last_error}>
                {status.last_error}
              </div>
            )}
          </div>
        </div>
        {config?.base_url && (
          <div className="mt-2 text-xs text-slate-500 truncate">
            {config.base_url}
          </div>
        )}
      </Card>

      {/* Foreground App */}
      <Card title="前台应用">
        {status.process_name ? (
          <div>
            <div className="text-sm font-semibold text-sky-300 truncate">
              {status.process_name}
            </div>
            {status.process_title && (
              <div className="text-xs text-slate-400 mt-1 truncate" title={status.process_title}>
                {status.process_title}
              </div>
            )}
          </div>
        ) : (
          <span className="text-sm text-slate-500">无</span>
        )}
      </Card>

      {/* Now Playing */}
      <Card title="正在播放">
        {hasMedia ? (
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 rounded-lg bg-slate-700 flex items-center justify-center text-xl flex-shrink-0">
              🎵
            </div>
            <div className="min-w-0">
              <div className="text-sm font-semibold text-white truncate">
                {status.media_title || "未知曲目"}
              </div>
              {status.media_artist && (
                <div className="text-xs text-slate-400 truncate mt-0.5">
                  {status.media_artist}
                </div>
              )}
              {status.media_album && (
                <div className="text-xs text-slate-500 truncate mt-0.5">
                  {status.media_album}
                </div>
              )}
            </div>
          </div>
        ) : (
          <span className="text-sm text-slate-500">无</span>
        )}
      </Card>

      {/* Device Info */}
      <Card title="设备信息">
        <div className="grid grid-cols-2 gap-2 text-xs">
          <div>
            <div className="text-slate-400">设备名</div>
            <div className="text-white mt-0.5 truncate">
              {config?.device_name || "（未设置）"}
            </div>
          </div>
          <div>
            <div className="text-slate-400">电量</div>
            <div className="mt-0.5">
              <BatteryIcon level={status.battery_level} charging={status.is_charging} />
            </div>
          </div>
        </div>
      </Card>
    </div>
  );
}
