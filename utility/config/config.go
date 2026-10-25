package config

import (
	"strconv"
	"time"

	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	"github.com/LandcLi/landc-go/log/facade"
)

// Config represents the application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Log      facade.LogConfig
	Org      OrgServiceConfig
	Cache    CacheConfig
}

// CacheConfig 控制进程内缓存的行为。
type CacheConfig struct {
	// EffectiveTTL 是「用户有效权限」缓存的存活时间。
	// 0 表示关闭缓存（排查「权限改了不生效」时先把它设为 0）。
	EffectiveTTL time.Duration
}

// defaultEffectiveCacheTTL 与 internal/service_impl/permission 的默认值保持一致。
const defaultEffectiveCacheTTL = 30 * time.Second

// OrgServiceConfig 描述与 LandcOrg 的集成方式（可选）。
// BaseURL 留空表示不集成：权限判定只使用直接分配的角色。
type OrgServiceConfig struct {
	BaseURL string
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Host             string
	Port             int
	GinMode          string
	EnableCORS       bool
	UserDefaultRoute bool
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

// RedisConfig represents Redis configuration.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// GetConfig 从框架全局配置中提取类型化配置。
// 配置由框架全权加载和管理，此函数仅做类型转换。
func GetConfig() *Config {
	fc := frameconfig.GetConfig()
	if fc == nil {
		panic("框架配置未初始化，请先调用 frameconfig.InitGlobalConfigWithPath")
	}

	return extractConfig(fc)
}

func extractConfig(fc *frameconfig.Config) *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host:             fc.Server.Addr,
			Port:             fc.Server.Port,
			UserDefaultRoute: fc.Server.UseDefaultRoutes,
		},
		Database: DatabaseConfig{
			DSN:             fc.Database.DSN,
			MaxOpenConns:    fc.Database.MaxOpenConns,
			MaxIdleConns:    fc.Database.MaxIdleConns,
			ConnMaxLifetime: fc.Database.ConnMaxLifetime,
		},
		Redis: RedisConfig{
			Addr:     fc.Redis.Addr,
			Password: fc.Redis.Password,
			DB:       fc.Redis.DB,
			PoolSize: fc.Redis.PoolSize,
		},
	}

	// 从自定义配置提取 GinMode / CORS / 组织服务地址。
	//
	// 读取统一走框架的 GetOtherString / GetOtherBool：它们按嵌套路径下钻，
	// 且框架在解析配置时会把 other: 段的子键提升一层（历史文档把业务配置
	// 写在 other 段下，而 Other 的语义是未声明的顶层键）。此前本包自己写的
	// 下钻逻辑只能处理其中一种写法，另一种会静默取空。
	cfg.Server.GinMode = fc.GetOtherString("server", "gin_mode")
	if b, ok := fc.GetOtherBool("server", "enable_cors"); ok {
		cfg.Server.EnableCORS = b
	} else if b, ok := fc.GetOtherBool("enable_cors"); ok {
		// 兼容顶层写法
		cfg.Server.EnableCORS = b
	}
	cfg.Org.BaseURL = fc.GetOtherString("org_service", "base_url")

	// 有效权限缓存存活时间（秒）。未配置用默认 30s；显式写 0 表示关闭缓存。
	cfg.Cache.EffectiveTTL = defaultEffectiveCacheTTL
	if v, ok := fc.GetOtherPath("permission", "effective_cache_ttl_seconds"); ok {
		cfg.Cache.EffectiveTTL = time.Duration(toInt(v)) * time.Second
	}

	// 从框架 LogConfig 转换为 facade.LogConfig
	cfg.Log = facade.LogConfig{
		Provider:    "console",
		Name:        "landc-permission",
		Level:       parseLogLevel(fc.Log.Level),
		Format:      fc.Log.Format,
		OutputPath:  fc.Log.Output,
		MaxLogSize:  int64(fc.Log.MaxSize),
		MaxLogFiles: fc.Log.MaxBackups,
		MaxLogAge:   time.Duration(fc.Log.MaxAge) * 24 * time.Hour,
	}

	return cfg
}

// toInt 把配置里读到的数值统一成 int。
// yaml 解码出 int，json 解码出 float64，两条路径都要能处理。
func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case uint:
		return int(n)
	case float64:
		return int(n)
	case string:
		if i, err := strconv.Atoi(n); err == nil {
			return i
		}
	}
	return 0
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) facade.LogLevel {
	switch level {
	case "debug":
		return facade.DebugLevel
	case "info":
		return facade.InfoLevel
	case "warn", "warning":
		return facade.WarnLevel
	case "error":
		return facade.ErrorLevel
	case "fatal":
		return facade.FatalLevel
	default:
		return facade.InfoLevel
	}
}
