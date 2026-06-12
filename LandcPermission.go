package landcpermission

import (
	"fmt"

	"github.com/LandcLi/LandcPermission/api"
	"github.com/LandcLi/LandcPermission/api/permission"
	"github.com/LandcLi/LandcPermission/api/role"
	_ "github.com/LandcLi/LandcPermission/internal"
	"github.com/LandcLi/LandcPermission/utility/config"
	"github.com/LandcLi/LandcPermission/utility/logger"
	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"github.com/LandcLi/landc-go/frame/pkg/middleware"
	"github.com/LandcLi/landc-go/frame/pkg/web"
	"github.com/gin-gonic/gin"
)

// InitDefault 从配置文件路径初始化，返回 web.Server。
// 框架全权管理配置加载，LPM 仅通过 config.GetConfig() 做类型化访问。
func InitDefault(configPath string) (*web.Server, error) {
	if err := frameconfig.InitGlobalConfigWithPathAndEnv(configPath); err != nil {
		return nil, fmt.Errorf("框架加载配置失败: %w", err)
	}
	return initLPM(config.GetConfig())
}

// InitWithConfig 从配置对象初始化，返回 web.Server
func InitWithConfig(cfg *config.Config) (*web.Server, error) {
	return initLPM(cfg)
}

// InitWithConfigAndGin 使用指定配置和已有 Gin 引擎初始化（兼容模式）
func InitWithConfigAndGin(cfg *config.Config, ginEngine *gin.Engine) error {
	return initLPMOnGin(cfg, ginEngine)
}

func initLPM(cfg *config.Config) (*web.Server, error) {
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	server := web.NewServer(&web.ServerConfig{
		Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
	})

	if err := initInfrastructure(cfg); err != nil {
		return nil, err
	}

	if cfg.Server.UserDefaultRoute {
		registerRoutes(server)
	}

	return server, nil
}

func initLPMOnGin(cfg *config.Config, ginEngine *gin.Engine) error {
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	if err := initInfrastructure(cfg); err != nil {
		return err
	}

	if cfg.Server.UserDefaultRoute {
		registerRoutesOnGin(ginEngine)
	}

	return nil
}

func initInfrastructure(cfg *config.Config) error {
	logger.InitLogger(&cfg.Log)

	framedb.InitGlobalDBWithDefault()
	framecache.InitGlobalCacheWithDefault()

	return nil
}

func registerRoutes(server *web.Server) {
	engine := server.Engine()
	engine.Use(middleware.Trace(), middleware.CORS(), middleware.Recovery(), middleware.Auth())
	api.RegisterDefaultRoutes(engine)

	// 注册控制器（自动从 meta.Meta 标签解析路由）
	server.RegisterHandler(role.RoleGateway.Get())
	server.RegisterHandler(permission.PermissionGateway.Get())
}

func registerRoutesOnGin(ginEngine *gin.Engine) {
	ginEngine.Use(middleware.Trace(), middleware.CORS(), middleware.Recovery())
	api.RegisterDefaultRoutes(ginEngine)

	// 创建临时 server 来注册控制器
	tempServer := web.NewServer(&web.ServerConfig{})
	tempServer.RegisterHandler(role.RoleGateway.Get())
	tempServer.RegisterHandler(permission.PermissionGateway.Get())

	// 复制路由到 ginEngine
	for _, route := range tempServer.Engine().Routes() {
		ginEngine.Handle(route.Method, route.Path, route.HandlerFunc)
	}
}
