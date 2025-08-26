package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds application configuration loaded from env and files.
type Config struct {
	Env  string `mapstructure:"APP_ENV"`
	Host string `mapstructure:"APP_HOST"`
	Port int    `mapstructure:"APP_PORT"`

	DB struct {
		Host     string `mapstructure:"DB_HOST"`
		Port     int    `mapstructure:"DB_PORT"`
		User     string `mapstructure:"DB_USER"`
		Password string `mapstructure:"DB_PASSWORD"`
		Name     string `mapstructure:"DB_NAME"`
		SSLMode  string `mapstructure:"DB_SSLMODE"`
	} `mapstructure:",squash"`

	CORS struct {
		AllowOrigins []string `mapstructure:"CORS_ALLOW_ORIGINS"`
		AllowMethods []string `mapstructure:"CORS_ALLOW_METHODS"`
		AllowHeaders []string `mapstructure:"CORS_ALLOW_HEADERS"`
	} `mapstructure:",squash"`
}

// Load loads the configuration with sane defaults and environment overrides.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_HOST", "0.0.0.0")
	v.SetDefault("APP_PORT", 8080)

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_USER", "volsync")
	v.SetDefault("DB_PASSWORD", "volsync")
	v.SetDefault("DB_NAME", "volsync")
	v.SetDefault("DB_SSLMODE", "disable")

	v.SetDefault("CORS_ALLOW_ORIGINS", []string{"*"})
	v.SetDefault("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("CORS_ALLOW_HEADERS", []string{"Authorization", "Content-Type"})

	// Load .env if present, ignore if missing
	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal: %w", err)
	}

	// Basic validation
	if cfg.Port == 0 {
		return nil, fmt.Errorf("APP_PORT must be > 0")
	}
	if cfg.DB.Host == "" || cfg.DB.User == "" || cfg.DB.Name == "" {
		return nil, fmt.Errorf("database configuration incomplete")
	}

	return &cfg, nil
}
