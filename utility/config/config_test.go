package config

import (
	"testing"

	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
)

// TestExtractConfigReadsNestedOther 锁定「other 段嵌套配置能被读到」这条约束。
//
// 历史缺陷：extractConfig 直接调用 frameconfig.Config.GetOther("gin_mode")，
// 而 GetOther 只按顶层 key 查表；配置文件里写的是 other.server.gin_mode，
// 因此取值恒为空 —— 配置写了不生效，且没有任何报错线索。
func TestExtractConfigReadsNestedOther(t *testing.T) {
	fc := &frameconfig.Config{
		Server: frameconfig.ServerConfig{Addr: "0.0.0.0", Port: 8082},
		Log:    frameconfig.LogConfig{Level: "info"},
		Other: map[string]interface{}{
			"server": map[string]interface{}{
				"gin_mode":    "release",
				"enable_cors": true,
			},
			"org_service": map[string]interface{}{
				"base_url": "http://127.0.0.1:8083",
			},
		},
	}

	cfg := extractConfig(fc)

	if cfg.Server.GinMode != "release" {
		t.Errorf("GinMode = %q, want %q", cfg.Server.GinMode, "release")
	}
	if !cfg.Server.EnableCORS {
		t.Error("EnableCORS = false, want true（来自 other.server.enable_cors）")
	}
	if cfg.Org.BaseURL != "http://127.0.0.1:8083" {
		t.Errorf("Org.BaseURL = %q, want %q", cfg.Org.BaseURL, "http://127.0.0.1:8083")
	}
}

// TestExtractConfigWithoutOther 验证未配置自定义段时的默认行为：
// 不得 panic，GinMode 留空（交给 gin 自身默认），Org 集成为关闭状态。
func TestExtractConfigWithoutOther(t *testing.T) {
	fc := &frameconfig.Config{
		Server: frameconfig.ServerConfig{Addr: "0.0.0.0", Port: 8082},
		Log:    frameconfig.LogConfig{Level: "error"},
	}

	cfg := extractConfig(fc)

	if cfg.Server.GinMode != "" {
		t.Errorf("GinMode = %q, want empty", cfg.Server.GinMode)
	}
	if cfg.Org.BaseURL != "" {
		t.Errorf("Org.BaseURL = %q, want empty（未配置即不集成）", cfg.Org.BaseURL)
	}
	if cfg.Server.Port != 8082 {
		t.Errorf("Port = %d, want 8082", cfg.Server.Port)
	}
}

// TestExtractConfigTypeMismatch 验证类型不符时安全忽略，而不是 panic。
func TestExtractConfigTypeMismatch(t *testing.T) {
	fc := &frameconfig.Config{
		Log: frameconfig.LogConfig{Level: "info"},
		Other: map[string]interface{}{
			"server": map[string]interface{}{
				"gin_mode":    123,   // 应为 string
				"enable_cors": "yes", // 应为 bool
			},
			"org_service": "not-a-map", // 应为 map
		},
	}

	cfg := extractConfig(fc)

	if cfg.Server.GinMode != "" {
		t.Errorf("GinMode = %q, want empty for non-string value", cfg.Server.GinMode)
	}
	if cfg.Server.EnableCORS {
		t.Error("EnableCORS = true, want false for non-bool value")
	}
	if cfg.Org.BaseURL != "" {
		t.Errorf("Org.BaseURL = %q, want empty for non-map section", cfg.Org.BaseURL)
	}
}
