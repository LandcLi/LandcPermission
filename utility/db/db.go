package db

import (
	framecache "github.com/LandcLi/landc-go/frame/pkg/cache"
	framedb "github.com/LandcLi/landc-go/frame/pkg/db"
	"gorm.io/gorm"
)

// 以下函数是对框架 db/cache 包的兼容性封装
// 新代码请直接使用 framedb.GetDB() 和 framecache.GetCache()

// GetDB 获取数据库连接（委托给框架）
func GetDB() *gorm.DB {
	return framedb.GetDB()
}

// GetRedis 获取 Redis 缓存实例（委托给框架 cache 包）
func GetRedis() any {
	return framecache.GetCache()
}
