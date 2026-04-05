import { useState } from "react";
import { Config } from "../App";

interface Props {
  config: Config;
  onSave: (cfg: Config) => void;
  saving: boolean;
  saveMsg: string | null;
}

interface FieldProps {
  label: string;
  description?: string;
  children: React.ReactNode;
}

function Field({ label, description, children }: FieldProps) {
  return (
    <div className="space-y-1.5">
      <label className="block text-sm font-medium text-slate-200">{label}</label>
      {description && (
        <p className="text-xs text-slate-400">{description}</p>
      )}
      {children}
    </div>
  );
}

const inputClass =
  "w-full bg-slate-800 border border-slate-600 rounded-lg px-3 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-sky-500 focus:border-transparent transition-colors";

interface RemoteConfig {
  endpoint?: string;
  apiKey?: string;
  token?: {
    reportEndpoint?: string;
    items?: { token?: string }[];
  };
}

export default function Settings({ config, onSave, saving, saveMsg }: Props) {
  const [form, setForm] = useState<Config>({ ...config });
  const [base64Input, setBase64Input] = useState("");
  const [base64Msg, setBase64Msg] = useState<string | null>(null);

  const set = (key: keyof Config, value: unknown) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(form);
  };

  const handleBase64Import = () => {
    const raw = base64Input.trim();
    if (!raw) return;
    try {
      const json: RemoteConfig = JSON.parse(atob(raw));
      const token =
        json.apiKey?.trim() ||
        json.token?.items?.[0]?.token?.trim() ||
        "";
      if (!token) {
        setBase64Msg("Base64 配置中未找到 apiKey 或 token");
        return;
      }
      const endpoint =
        json.endpoint?.trim() ||
        json.token?.reportEndpoint?.trim() ||
        "";
      let baseUrl = "http://localhost:3000";
      if (endpoint) {
        baseUrl = endpoint.replace(/\/api\/activity$/, "").replace(/\/$/, "") || baseUrl;
      }
      setForm((prev) => ({ ...prev, base_url: baseUrl, api_token: token }));
      setBase64Msg("导入成功，请保存配置");
    } catch {
      setBase64Msg("Base64 配置解析失败，请检查内容是否正确");
    }
  };

  return (
    <form onSubmit={handleSubmit} className="p-4 space-y-5">
      <Field
        label="从 Base64 导入配置"
        description="粘贴后台一键导出的 Base64 配置，自动填充地址和 Token"
      >
        <div className="flex gap-2">
          <input
            type="text"
            className={inputClass}
            placeholder="粘贴 Base64 配置..."
            value={base64Input}
            onChange={(e) => {
              setBase64Input(e.target.value);
              setBase64Msg(null);
            }}
          />
          <button
            type="button"
            onClick={handleBase64Import}
            className="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-white text-sm font-medium rounded-lg transition-colors whitespace-nowrap"
          >
            导入
          </button>
        </div>
        {base64Msg && (
          <p
            className={`text-xs mt-1 ${
              base64Msg.startsWith("导入成功") ? "text-green-400" : "text-red-400"
            }`}
          >
            {base64Msg}
          </p>
        )}
      </Field>

      <Field
        label="后端地址"
        description="Waken-Wa 后端的根地址（不含末尾 /）"
      >
        <input
          type="text"
          className={inputClass}
          placeholder="http://localhost:3000"
          value={form.base_url}
          onChange={(e) => set("base_url", e.target.value)}
        />
      </Field>

      <Field
        label="API Token"
        description="Bearer Token，用于鉴权"
      >
        <input
          type="password"
          className={inputClass}
          placeholder="wk_..."
          value={form.api_token}
          onChange={(e) => set("api_token", e.target.value)}
        />
      </Field>

      <Field
        label="设备名"
        description="显示在后端的设备名称（留空则使用主机名）"
      >
        <input
          type="text"
          className={inputClass}
          placeholder="My Desktop"
          value={form.device_name}
          onChange={(e) => set("device_name", e.target.value)}
        />
      </Field>

      <Field
        label="设备 Hash Key"
        description="稳定的设备唯一标识，首次运行自动生成"
      >
        <input
          type="text"
          className={inputClass}
          placeholder="（自动生成）"
          value={form.generated_hash_key}
          onChange={(e) => set("generated_hash_key", e.target.value)}
        />
      </Field>

      <div className="grid grid-cols-2 gap-4">
        <Field
          label="轮询间隔（ms）"
          description="默认 2000ms"
        >
          <input
            type="number"
            className={inputClass}
            placeholder="2000"
            min={100}
            value={form.poll_interval_ms ?? ""}
            onChange={(e) =>
              set("poll_interval_ms", e.target.value ? parseInt(e.target.value) : null)
            }
          />
        </Field>

        <Field
          label="心跳间隔（ms）"
          description="0 = 关闭；默认 60000ms"
        >
          <input
            type="number"
            className={inputClass}
            placeholder="60000"
            min={0}
            value={form.heartbeat_interval_ms ?? ""}
            onChange={(e) =>
              set("heartbeat_interval_ms", e.target.value ? parseInt(e.target.value) : null)
            }
          />
        </Field>
      </div>

      <Field label="绕过系统代理">
        <label className="flex items-center gap-3 cursor-pointer">
          <div
            onClick={() => set("bypass_system_proxy", !form.bypass_system_proxy)}
            className={`relative w-11 h-6 rounded-full transition-colors cursor-pointer ${
              form.bypass_system_proxy ? "bg-sky-600" : "bg-slate-600"
            }`}
          >
            <div
              className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform ${
                form.bypass_system_proxy ? "translate-x-5" : "translate-x-0"
              }`}
            />
          </div>
          <span className="text-sm text-slate-300">
            {form.bypass_system_proxy ? "已启用" : "已禁用"}
          </span>
        </label>
      </Field>

      <div className="pt-2 flex items-center gap-3">
        <button
          type="submit"
          disabled={saving}
          className="px-5 py-2 bg-sky-600 hover:bg-sky-500 disabled:bg-slate-600 text-white text-sm font-medium rounded-lg transition-colors"
        >
          {saving ? "保存中..." : "保存配置"}
        </button>
        {saveMsg && (
          <span
            className={`text-sm ${
              saveMsg.startsWith("配置") ? "text-green-400" : "text-red-400"
            }`}
          >
            {saveMsg}
          </span>
        )}
      </div>
    </form>
  );
}
