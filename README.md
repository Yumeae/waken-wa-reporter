## Waken-Wa-Reporter

适用于 [Waken-Wa](https://github.com/MoYoez/Waken-Wa) 的桌面端 **Activity Reporter**：周期性采集当前前台应用与（在支持的系统上）正在播放的媒体信息，并上报到 Waken-Wa 后端 `POST /api/activity`。

也有适用于 Auto.js 的 **Android** 版本（能力见下表）。

## 平台兼容

以下为 **官方支持** 的能力范围（与实现一致）。

| 能力 | Windows | macOS | Android |
|------|:-------:|:-----:|:-------:|
| 程序名 / 前台进程名（`process_name`） | ✅ | ✅ | ✅ |
| 窗口标题（`process_title`） | ✅ | ❌ | ❌ |
| 正在播放（`metadata.media`） | ✅ | ❌ | ❌ |
| 电量（battery） | ❌ | ❌ | ✅ |

说明：

- **Windows**：前台进程名、窗口标题均支持。**正在播放**（`metadata.media`）依赖 **SMTC**（System Media Transport Controls，系统媒体传输控件）：播放器需向系统媒体会话公开元数据（例如多数现代浏览器、Spotify、系统「正在播放」浮层能显示的应用）；未接入 SMTC 的应用无法提供曲目信息。
- **macOS**：仅支持 **前台进程 / 应用名**；窗口标题留空；不包含「正在播放」媒体信息。
- **Android**（Auto.js 等）：支持 **当前前台应用名**（`process_name`）与 **电量** 上报；窗口标题与系统级「正在播放」媒体元数据不在此列能力内。


## 环境要求

- [Go](https://go.dev/dl/) 1.25+

## 构建与运行

在项目根目录执行：

```bash
go build -o waken-wa-reporter .
```

首次在终端中交互式写入配置（API 地址、Token、设备名、轮询/心跳间隔等）：

```bash
./waken-wa-reporter -setup
```

日常启动（会读取已保存配置或环境变量）：

```bash
./waken-wa-reporter
```

若未设置 Token 且无可用配置文件，且标准输入非 TTY（例如作为服务运行），程序会退出并提示先设置 `WAKEN_API_TOKEN` 或完成本地配置。

## 配置方式

优先级（由高到低）大致为：**环境变量** → **保存的 `config.json`** → **交互式引导**（仅在有 TTY 且未配置时）。

### 配置文件路径

默认路径为操作系统用户配置目录下的 `waken-wa/config.json`，例如：

- Windows：`%APPDATA%\waken-wa\config.json`
- macOS：`$XDG_CONFIG_HOME/waken-wa/config.json`（未设置时多为 `~/.config/waken-wa/config.json`）

### 常用环境变量

| 变量 | 说明 |
|------|------|
| `WAKEN_API_TOKEN` | API Token（与后台一致） |
| `WAKEN_BASE_URL` | 后端根地址，默认 `http://localhost:3000` |
| `WAKEN_CONFIG_BASE64` | 后台导出的一键 Base64 JSON，可解析出 URL 与 Token |
| `WAKEN_DEVICE` | 上报中的设备标识（默认：配置中的设备名，或主机名） |
| `WAKEN_DEVICE_NAME` | 展示用设备名 |
| `WAKEN_GENERATED_HASH_KEY` | 稳定哈希键；未设置时会在首次运行生成并写入配置文件 |
| `WAKEN_POLL_INTERVAL` | 轮询间隔，Go `time.ParseDuration` 格式（如 `2s`） |
| `WAKEN_HEARTBEAT_INTERVAL` | 心跳间隔；`0` 表示关闭心跳 |
| `WAKEN_DEVICE_TYPE` | `desktop` / `tablet` / `mobile`（默认 `desktop`） |
| `WAKEN_PUSH_MODE` | `realtime` 或 `active`（默认 `realtime`） |
| `WAKEN_BATTERY_LEVEL` | 可选 `0`–`100` |
| `WAKEN_METADATA` | 额外 JSON 元数据，会合并进上报 `metadata` |

## 上报说明

客户端向 `{WAKEN_BASE_URL}/api/activity` 发送 JSON，包含 `generated_hash_key`、`device`、`process_name`、`process_title`、可选 `metadata`（含 `source: waken-wa` 与媒体信息等）。需使用有效的 Bearer Token。

## License

MIT
