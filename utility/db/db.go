package db

import (
	"github.com/LandcLi/LandcPermission/utility/config"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	database    *gorm.DB
	redisClient *redis.Client
)

// InitDatabase initializes the database connection.
func InitDatabase(cfg *config.DatabaseConfig) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	database = db
}

// GetDB returns the database instance.
func GetDB() *gorm.DB {
	return database
}

// InitRedis initializes the Redis connection (optional).
// If cfg is nil or Addr is empty, Redis will not be initialized.
func InitRedis(cfg *config.RedisConfig) {
	if cfg == nil || cfg.Addr == "" {
		return
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

// GetRedis returns the Redis client (may be nil if not configured).
func GetRedis() *redis.Client {
	return redisClient
}
