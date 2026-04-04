package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/MoYoez/waken-wa-reporter/internal/activity"
	"github.com/MoYoez/waken-wa-reporter/internal/cliutil"
	"github.com/MoYoez/waken-wa-reporter/internal/config"
	"github.com/MoYoez/waken-wa-reporter/internal/platform/foreground"
	"github.com/MoYoez/waken-wa-reporter/internal/platform/media"
	"github.com/MoYoez/waken-wa-reporter/internal/platform/power"
	"golang.org/x/term"
)

type stringListFlag []string

func (s *stringListFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringListFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func printConfigFile(cfg *config.File) error {
	if cfg == nil {
		cfg = &config.File{}
	}
	body, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", body)
	return nil
}

func applyConfigUpdate(cfg *config.File, update string) error {
	key, value, ok := strings.Cut(update, "=")
	if !ok {
		return fmt.Errorf("配置项格式错误：%s（应为 key=value）", update)
	}

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" {
		return fmt.Errorf("配置项格式错误：%s（key 不能为空）", update)
	}

	switch key {
	case "base_url":
		if value == "" {
			cfg.BaseURL = ""
			return nil
		}
		cfg.BaseURL = strings.TrimRight(value, "/")
		return nil
	case "api_token":
		cfg.APIToken = value
		return nil
	case "device_name":
		cfg.DeviceName = value
		return nil
	case "generated_hash_key":
		cfg.GeneratedHashKey = value
		return nil
	case "poll_interval_ms":
		if value == "" {
			cfg.PollIntervalMs = nil
			return nil
		}
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 {
			return fmt.Errorf("poll_interval_ms 必须是 >=1 的整数毫秒")
		}
		cfg.PollIntervalMs = &n
		return nil
	case "heartbeat_interval_ms":
		if value == "" {
			cfg.HeartbeatIntervalMs = nil
			return nil
		}
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return fmt.Errorf("heartbeat_interval_ms 必须是 >=0 的整数毫秒")
		}
		cfg.HeartbeatIntervalMs = &n
		return nil
	case "metadata":
		if value == "" {
			cfg.Metadata = nil
			return nil
		}
		var metadata map[string]any
		if err := json.Unmarshal([]byte(value), &metadata); err != nil {
			return fmt.Errorf("metadata 必须是 JSON object: %w", err)
		}
		cfg.Metadata = metadata
		return nil
	case "bypass_system_proxy":
		if value == "" {
			cfg.BypassSystemProxy = false
			return nil
		}
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("bypass_system_proxy 必须是 true/false/1/0")
		}
		cfg.BypassSystemProxy = b
		return nil
	default:
		return fmt.Errorf("不支持的配置项：%s", key)
	}
}

// formatMediaForLog returns a short single-line summary for logs, or "" if nothing to show.
func formatMediaForLog(m media.Info) string {
	if m.IsEmpty() {
		return ""
	}
	var parts []string
	if t := strings.TrimSpace(m.Title); t != "" {
		parts = append(parts, t)
	}
	if a := strings.TrimSpace(m.Artist); a != "" {
		parts = append(parts, a)
	}
	if al := strings.TrimSpace(m.Album); al != "" {
		parts = append(parts, al)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " — ")
}

// openBrowser opens url in the system default handler (best-effort).
func openBrowser(url string) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("打开浏览器失败：URL 为空")
	}
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func resolveApprovalRetryInterval() time.Duration {
	s := strings.TrimSpace(os.Getenv("WAKEN_APPROVAL_RETRY_INTERVAL"))
	if s == "" {
		return 45 * time.Second
	}
	d, err := time.ParseDuration(s)
	if err != nil || d < 5*time.Second {
		return 45 * time.Second
	}
	return d
}

func maybeWriteApprovalURLFile(url string) {
	path := strings.TrimSpace(os.Getenv("WAKEN_APPROVAL_URL_FILE"))
	if path == "" || url == "" {
		return
	}
	if err := os.WriteFile(path, []byte(url+"\n"), 0o600); err != nil {
		log.Printf("审核：无法写入 %s：%v", path, err)
	}
}

func maybeOpenApprovalURL(url string) {
	if strings.TrimSpace(os.Getenv("WAKEN_OPEN_APPROVAL")) != "1" {
		return
	}
	if err := openBrowser(url); err != nil {
		log.Printf("审核：无法打开浏览器：%v", err)
	}
}

func main() {
	setup := flag.Bool("setup", false, "run interactive setup (URL + API token), save, and exit")
	printConfig := flag.Bool("print-config", false, "print saved config JSON and exit")
	var setConfig stringListFlag
	flag.Var(&setConfig, "set-config", "update saved config with key=value and exit; repeatable")
	flag.Parse()

	if *printConfig && len(setConfig) > 0 {
		log.Fatal("-print-config 与 -set-config 不能同时使用")
	}

	if *printConfig || len(setConfig) > 0 {
		path, err := config.DefaultFilePath()
		if err != nil {
			log.Fatal(err)
		}

		cfg, err := config.Load(path)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatal(err)
			}
			cfg = &config.File{}
		}

		if *printConfig {
			if err := printConfigFile(cfg); err != nil {
				log.Fatal(err)
			}
			return
		}

		for _, update := range setConfig {
			if err := applyConfigUpdate(cfg, update); err != nil {
				log.Fatal(err)
			}
		}
		if err := config.Save(path, cfg); err != nil {
			log.Fatal(err)
		}
		if err := printConfigFile(cfg); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *setup {
		path, err := config.DefaultFilePath()
		if err != nil {
			log.Fatal(err)
		}
		if _, _, err := config.RunWizard(path); err != nil {
			log.Fatal(err)
		}
		return
	}

	baseURL, token, err := config.Resolve()
	if err != nil {
		log.Fatal(err)
	}

	deviceName, err := config.ResolveDeviceName()
	if err != nil {
		log.Fatalf("解析设备名称失败：%v", err)
	}
	generatedHashKey, err := config.ResolveGeneratedHashKey()
	if err != nil {
		log.Fatalf("解析设备身份牌失败：%v", err)
	}

	device := strings.TrimSpace(os.Getenv("WAKEN_DEVICE"))
	if device == "" {
		if deviceName != "" {
			device = deviceName
		} else {
			h, err := os.Hostname()
			if err != nil {
				log.Fatalf("读取主机名失败：%v", err)
			}
			device = h
		}
	}

	if deviceName == "" {
		deviceName = device
	}

	poll, err := config.ResolvePollInterval()
	if err != nil {
		log.Fatalf("解析轮询间隔失败：%v", err)
	}

	heartbeat, heartbeatEnabled, err := config.ResolveHeartbeatInterval()
	if err != nil {
		log.Fatalf("解析心跳间隔失败：%v", err)
	}

	meta := map[string]any{"source": "waken-wa"}
	if path, err := config.DefaultFilePath(); err == nil {
		if f, err := config.Load(path); err == nil && f.Metadata != nil {
			activity.MergeMetadata(meta, f.Metadata)
		}
	}
	if s := os.Getenv("WAKEN_METADATA"); s != "" {
		var extra map[string]any
		if err := json.Unmarshal([]byte(s), &extra); err != nil {
			log.Fatalf("解析 WAKEN_METADATA 失败：%v", err)
		}
		activity.MergeMetadata(meta, extra)
	}

	deviceType := strings.TrimSpace(os.Getenv("WAKEN_DEVICE_TYPE"))
	if deviceType == "" {
		deviceType = "desktop"
	}
	if deviceType != "desktop" && deviceType != "tablet" && deviceType != "mobile" {
		log.Fatalf("WAKEN_DEVICE_TYPE 只能是 desktop/tablet/mobile，当前为：%s", deviceType)
	}

	pushMode := strings.TrimSpace(os.Getenv("WAKEN_PUSH_MODE"))
	if pushMode == "" {
		pushMode = "realtime"
	}
	if pushMode != "realtime" && pushMode != "active" {
		log.Fatalf("WAKEN_PUSH_MODE 只能是 realtime/active，当前为：%s", pushMode)
	}

	var batteryLevel *int
	if s := strings.TrimSpace(os.Getenv("WAKEN_BATTERY_LEVEL")); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			log.Fatalf("解析 WAKEN_BATTERY_LEVEL 失败：%v", err)
		}
		if v < 0 || v > 100 {
			log.Fatalf("WAKEN_BATTERY_LEVEL 必须在 [0,100] 范围内，当前为：%d", v)
		}
		batteryLevel = &v
	}

	isCharging := power.IsCharging()

	bypassProxy, err := config.ResolveBypassSystemProxy()
	if err != nil {
		log.Fatalf("读取配置失败：%v", err)
	}
	client := &activity.Client{BaseURL: baseURL, Token: token}
	if bypassProxy {
		client.HTTPClient = activity.HTTPClientBypassProxy()
	}
	approvalEvery := resolveApprovalRetryInterval()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	type lastState struct {
		snap     foreground.Snapshot
		mediaSig string
	}
	var last *lastState
	var lastReport time.Time

	var pendingMode bool
	var lastPendingRetry time.Time
	var bannerShown bool

	enterPending := func(p *activity.PendingApprovalError) {
		pendingMode = true
		lastPendingRetry = time.Now()
		if p.Message != "" {
			log.Printf("审核：%s", p.Message)
		}
		if !bannerShown {
			if term.IsTerminal(int(os.Stdout.Fd())) {
				cliutil.PrintApprovalBanner(p.ApprovalURL)
			} else {
				log.Printf("审核链接（非终端环境）：%s", p.ApprovalURL)
			}
			bannerShown = true
		}
		maybeWriteApprovalURLFile(p.ApprovalURL)
		maybeOpenApprovalURL(p.ApprovalURL)
	}

	postReport := func(snap foreground.Snapshot, minfo media.Info, merr error, heartbeat bool) error {
		reportMeta := maps.Clone(meta)
		if merr == nil && !minfo.IsEmpty() {
			activity.MergeMetadata(reportMeta, map[string]any{"media": minfo.AsMap()})
			if v, ok := reportMeta["play_source"]; !ok || strings.TrimSpace(fmt.Sprint(v)) == "" {
				if sourceAppID := strings.TrimSpace(minfo.SourceAppID); sourceAppID != "" {
					reportMeta["play_source"] = sourceAppID
				} else {
					reportMeta["play_source"] = "system_media"
				}
			}
		} else if merr != nil && !errors.Is(merr, media.ErrNoMedia) && !errors.Is(merr, media.ErrUnsupported) {
			log.Printf("媒体信息：%v", merr)
		}
		err := client.Post(ctx, activity.ReportRequest{
			GeneratedHashKey: generatedHashKey,
			Device:           device,
			DeviceName:       deviceName,
			DeviceType:       deviceType,
			ProcessName:      snap.ProcessName,
			ProcessTitle:     snap.ProcessTitle,
			BatteryLevel:     batteryLevel,
			IsCharging:       isCharging,
			PushMode:         pushMode,
			Metadata:         reportMeta,
		})
		if err != nil {
			var p *activity.PendingApprovalError
			if errors.As(err, &p) {
				return err
			}
			log.Printf("上报失败：%v", err)
			return err
		}
		mediaSuffix := ""
		if merr == nil {
			if s := formatMediaForLog(minfo); s != "" {
				mediaSuffix = " | 媒体： " + s
			}
		}
		if heartbeat {
			log.Printf("活动心跳：%s%s", snap.ProcessName, mediaSuffix)
		} else {
			log.Printf("活动已上报：%s%s", snap.ProcessName, mediaSuffix)
		}
		return nil
	}

	mediaSignature := func(minfo media.Info, merr error) string {
		if merr != nil || minfo.IsEmpty() {
			return ""
		}
		return minfo.Signature()
	}

	if snap, err := foreground.GetSnapshot(); err != nil {
		log.Printf("前台应用：%v", err)
	} else {
		minfo, merr := media.GetNowPlaying()
		sig := mediaSignature(minfo, merr)
		last = &lastState{snap: snap, mediaSig: sig}
		err := postReport(snap, minfo, merr, false)
		var p *activity.PendingApprovalError
		if errors.As(err, &p) {
			enterPending(p)
		} else if err == nil {
			lastReport = time.Now()
		}
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("正在退出")
			return
		case <-ticker.C:
			if pendingMode {
				if time.Since(lastPendingRetry) < approvalEvery {
					continue
				}
				lastPendingRetry = time.Now()
				snap, err := foreground.GetSnapshot()
				if err != nil {
					log.Printf("前台应用：%v", err)
					continue
				}
				minfo, merr := media.GetNowPlaying()
				err = postReport(snap, minfo, merr, false)
				var p *activity.PendingApprovalError
				if errors.As(err, &p) {
					continue
				}
				if err != nil {
					continue
				}
				pendingMode = false
				log.Println("设备已通过审核，恢复正常上报")
				last = &lastState{snap: snap, mediaSig: mediaSignature(minfo, merr)}
				lastReport = time.Now()
				continue
			}

			snap, err := foreground.GetSnapshot()
			if err != nil {
				log.Printf("前台应用：%v", err)
				continue
			}
			minfo, merr := media.GetNowPlaying()
			sig := mediaSignature(minfo, merr)
			same := last != nil &&
				last.snap.ProcessName == snap.ProcessName &&
				last.snap.ProcessTitle == snap.ProcessTitle &&
				last.mediaSig == sig
			if same {
				if heartbeatEnabled && time.Since(lastReport) >= heartbeat {
					err := postReport(snap, minfo, merr, true)
					var p *activity.PendingApprovalError
					if errors.As(err, &p) {
						enterPending(p)
						continue
					}
					if err == nil {
						lastReport = time.Now()
					}
				}
				continue
			}
			last = &lastState{snap: snap, mediaSig: sig}
			err = postReport(snap, minfo, merr, false)
			var p *activity.PendingApprovalError
			if errors.As(err, &p) {
				enterPending(p)
				continue
			}
			if err == nil {
				lastReport = time.Now()
			}
		}
	}
}
