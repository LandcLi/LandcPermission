package landcpermission

import (
	"github.com/LandcLi/LandcPermission/api"
	_ "github.com/LandcLi/LandcPermission/internal"
	"github.com/LandcLi/LandcPermission/utility/config"
	"github.com/LandcLi/LandcPermission/utility/db"
	"github.com/LandcLi/LandcPermission/utility/logger"
	apiMiddleware "github.com/LandcLi/landc-go/api/middleware/gin"
	ginLogger "github.com/LandcLi/landc-go/log/adapter/gin"
	"github.com/gin-gonic/gin"
)

// InitDefault initializes from config file path.
func InitDefault(configPath string) (*gin.Engine, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	ginEngine := gin.Default()

	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	if cfg.Server.EnableCORS {
		ginEngine.Use(corsMiddleware())
	}

	logger.InitLogger(&cfg.Log)
	db.InitRedis(&cfg.Redis)
	db.InitDatabase(&cfg.Database)

	ginEngine.Use(apiMiddleware.GinMiddleware())
	ginLogger.UseWithGin(ginEngine, logger.GetLogger())

	if cfg.Server.UseDefaultRoute {
		api.RegisterDefaultRoutes(ginEngine)
	}

	return ginEngine, nil
}

// InitWithConfig initializes from config object.
func InitWithConfig(cfg *config.Config) (*gin.Engine, error) {
	ginEngine := gin.Default()

	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	if cfg.Server.EnableCORS {
		ginEngine.Use(corsMiddleware())
	}

	logger.InitLogger(&cfg.Log)
	db.InitRedis(&cfg.Redis)
	db.InitDatabase(&cfg.Database)

	ginEngine.Use(apiMiddleware.GinMiddleware())
	ginLogger.UseWithGin(ginEngine, logger.GetLogger())

	if cfg.Server.UseDefaultRoute {
		api.RegisterDefaultRoutes(ginEngine)
	}

	return ginEngine, nil
}

// InitWithConfigAndGin initializes with existing Gin engine.
func InitWithConfigAndGin(cfg *config.Config, ginEngine *gin.Engine) error {
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	if cfg.Server.EnableCORS {
		ginEngine.Use(corsMiddleware())
	}

	logger.InitLogger(&cfg.Log)
	db.InitRedis(&cfg.Redis)
	db.InitDatabase(&cfg.Database)

	ginEngine.Use(apiMiddleware.GinMiddleware())
	ginLogger.UseWithGin(ginEngine, logger.GetLogger())

	if cfg.Server.UseDefaultRoute {
		api.RegisterDefaultRoutes(ginEngine)
	}

	return nil
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
