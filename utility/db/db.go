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

// InitRedis initializes the Redis connection.
func InitRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return redisClient, nil
}

// GetRedis returns the Redis client.
func GetRedis() *redis.Client {
	return redisClient
}
