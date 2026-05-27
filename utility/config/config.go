package config

import (
	"os"

	"github.com/LandcLi/landc-go/log/facade"
	"gopkg.in/yaml.v2"
)

// Config represents the application configuration.
type Config struct {
	Server   ServerConfig     `yaml:"server"`
	Database DatabaseConfig   `yaml:"database"`
	Redis    RedisConfig      `yaml:"redis"`
	Log      facade.LogConfig `yaml:"log"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Port            string `yaml:"port"`
	GinMode         string `yaml:"gin_mode"`
	EnableCORS      bool   `yaml:"enable_cors"`
	UseDefaultRoute bool   `yaml:"use_default_route"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

// RedisConfig represents Redis configuration.
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
