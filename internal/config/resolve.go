package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Resolve returns base URL and API token. Priority: env WAKEN_* > saved file > interactive wizard.
func Resolve() (baseURL, token string, err error) {
	token = strings.TrimSpace(os.Getenv("WAKEN_API_TOKEN"))
	baseURL = strings.TrimSpace(os.Getenv("WAKEN_BASE_URL"))
	if token != "" {
		if baseURL == "" {
			baseURL = defaultBaseURL
		}
		return strings.TrimRight(baseURL, "/"), token, nil
	}

	path, err := DefaultFilePath()
	if err != nil {
		return "", "", err
	}
	if f, err := Load(path); err == nil && strings.TrimSpace(f.APIToken) != "" {
		u := EffectiveBaseURL(f)
		if envURL := strings.TrimSpace(os.Getenv("WAKEN_BASE_URL")); envURL != "" {
			u = strings.TrimRight(envURL, "/")
		}
		return u, strings.TrimSpace(f.APIToken), nil
	}

	if !isCharDevice(os.Stdin) {
		return "", "", errors.New("未设置 WAKEN_API_TOKEN 且无已保存配置：请设置环境变量，或在终端中首次运行以完成引导")
	}

	return RunWizard(path)
}

func isCharDevice(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// RunWizard prompts for API base URL and token, then saves to savePath.
func RunWizard(savePath string) (baseURL, token string, err error) {
	fmt.Println()
	fmt.Println("  waken-wa — 首次配置")
	fmt.Println("  请填写 API 地址与 Token（后台 /admin → API Token）。")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("  API 地址 [%s]: ", defaultBaseURL)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	baseURL = strings.TrimSpace(strings.TrimRight(line, "\r\n"))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	fmt.Print("  API Token: ")
	line, err = reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	token = strings.TrimSpace(strings.TrimRight(line, "\r\n"))
	if token == "" {
		return "", "", errors.New("必须填写 API Token")
	}

	f := &File{BaseURL: baseURL, APIToken: token}
	if err := Save(savePath, f); err != nil {
		fmt.Fprintf(os.Stderr, "  警告：无法保存配置：%v\n", err)
	} else {
		fmt.Printf("\n  已保存到 %s\n\n", savePath)
	}

	return baseURL, token, nil
}
