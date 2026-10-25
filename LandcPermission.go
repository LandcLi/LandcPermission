package landcpermission

import (
	"fmt"

	"github.com/gin-gonic/gin"

	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	frameconfig "github.com/LandcLi/landc-go/frame/pkg/config"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"github.com/LandcLi/landc-go/frame/pkg/middleware"
	"github.com/LandcLi/landc-go/frame/pkg/web"

	"github.com/LandcLi/LandcPermission/api/permission"
	"github.com/LandcLi/LandcPermission/api/role"
	"github.com/LandcLi/LandcPermission/integration/orgrole"
	_ "github.com/LandcLi/LandcPermission/internal" // 触发默认实现注册（DI）
	permissionsvc "github.com/LandcLi/LandcPermission/internal/service_impl/permission"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/service"
	"github.com/LandcLi/LandcPermission/utility/config"
	"github.com/LandcLi/LandcPermission/utility/logger"
)

// InitDefault 从配置文件路径初始化，返回 web.Server。
// 框架全权管理配置加载，LPM 仅通过 config.GetConfig() 做类型化访问。
func InitDefault(configPath string) (*web.Server, error) {
	if err := frameconfig.InitGlobalConfigWithPathAndEnv(configPath); err != nil {
		return nil, fmt.Errorf("框架加载配置失败: %w", err)
	}
	return InitWithConfig(config.GetConfig())
}

// InitWithConfig 从配置对象初始化，返回 web.Server。
func InitWithConfig(cfg *config.Config) (*web.Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	server := web.NewServer(&web.ServerConfig{
		Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
	})

	// 框架级中间件（trace / CORS / recovery / 鉴权）
	server.Engine().Use(middleware.Trace(), middleware.CORS(), middleware.Recovery(), middleware.Auth())

	if err := initInfrastructure(cfg); err != nil {
		return nil, err
	}

	if cfg.Server.UserDefaultRoute {
		if err := RegisterToRouter(server); err != nil {
			return nil, err
		}
	}
	return server, nil
}

// InitWithConfigAndGin 使用已有 Gin 引擎初始化（库模式 / 嵌入到宿主）。
func InitWithConfigAndGin(cfg *config.Config, ginEngine *gin.Engine) error {
	if cfg == nil {
		return fmt.Errorf("配置不能为空")
	}
	if ginEngine == nil {
		return fmt.Errorf("gin 引擎不能为空")
	}
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	ginEngine.Use(middleware.Trace(), middleware.CORS(), middleware.Recovery())

	if err := initInfrastructure(cfg); err != nil {
		return err
	}
	if cfg.Server.UserDefaultRoute {
		return RegisterHandlers(ginEngine)
	}
	return nil
}

// RegisterToRouter 把本系统的控制器注册到 server（独立部署使用）。
func RegisterToRouter(server *web.Server) error {
	if err := server.RegisterHandler(role.GetRoleController()); err != nil {
		return fmt.Errorf("注册 role 控制器失败: %w", err)
	}
	if err := server.RegisterHandler(permission.GetPermissionController()); err != nil {
		return fmt.Errorf("注册 permission 控制器失败: %w", err)
	}
	return nil
}

// RegisterHandlers 把本系统的控制器注册到任意 gin 路由（嵌入到宿主时使用）。
// 不会挂载全局中间件：中间件链由宿主决定。
func RegisterHandlers(router gin.IRouter) error {
	if _, err := web.RegisterHandlers(router, role.GetRoleController()); err != nil {
		return fmt.Errorf("注册 role 控制器失败: %w", err)
	}
	if _, err := web.RegisterHandlers(router, permission.GetPermissionController()); err != nil {
		return fmt.Errorf("注册 permission 控制器失败: %w", err)
	}
	return nil
}

// initInfrastructure 初始化日志 / DB / 缓存 / 表结构。
// 前置条件：调用方已完成框架配置加载（InitDefault 或宿主 bootstrap）。
func initInfrastructure(cfg *config.Config) error {
	logger.InitLogger(&cfg.Log)

	if err := framedb.InitGlobalDBWithDefault(); err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	if err := framecache.InitGlobalCacheWithDefault(); err != nil {
		return fmt.Errorf("初始化缓存失败: %w", err)
	}
	if err := EnsureSchema(); err != nil {
		return err
	}
	wireOrgService(cfg)
	wirePermissionCache(cfg)
	return nil
}

// wirePermissionCache 按配置设定「有效权限缓存」的存活时间。
// 设为 0 即关闭缓存，排查「权限改了不生效」时先关它。
func wirePermissionCache(cfg *config.Config) {
	if cfg == nil {
		return
	}
	permissionsvc.SetEffectiveCacheTTL(cfg.Cache.EffectiveTTL)
}

// wireOrgService 按配置接入 LandcOrg 的组织继承角色（可选能力）。
//
// 未配置 other.org_service.base_url 时保持不集成：权限判定只用直接分配的
// 角色与直接权限，行为与未引入本能力前一致。配置后，有效权限会并入用户从
// 部门 / 岗位继承的角色；LandcOrg 不可用时自动降级为只用直接角色
// （见 internal/service_impl/permission 的 getAllUserRoleIDs）。
func wireOrgService(cfg *config.Config) {
	if cfg == nil || cfg.Org.BaseURL == "" {
		return
	}
	// 用 Override 而非 Provide：重复初始化（测试、宿主热重启）时替换旧实例，
	// 避免残留指向旧 base_url 的 provider。
	service.OverrideOrgRoleProvider(orgrole.NewRemoteProvider(cfg.Org.BaseURL))
}

// Models 返回本系统负责的持久化模型（模型清单唯一来源）。
func Models() []interface{} {
	return []interface{}{
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.UserRole{},
		&model.UserPermission{},
	}
}

// EnsureSchema 按配置决定是否自动建表（受 database.auto_migrate 控制），幂等。
// SQLite / 开发测试环境开启即可开箱使用；MySQL 生产环境由 sqls/init.sql 接管。
func EnsureSchema() error {
	if err := framedb.AutoMigrateIfEnabled(Models()...); err != nil {
		return fmt.Errorf("自动建表失败: %w", err)
	}
	return nil
}
