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

	Uploads struct {
		BaseDir string `mapstructure:"UPLOADS_BASE_DIR"`
		BaseURL string `mapstructure:"UPLOADS_BASE_URL"`
		MaxMB   int    `mapstructure:"UPLOADS_MAX_MB"`
	} `mapstructure:",squash"`

	JWT struct {
		AccessSecret   string `mapstructure:"JWT_ACCESS_SECRET"`
		RefreshSecret  string `mapstructure:"JWT_REFRESH_SECRET"`
		AccessTTLMin   int    `mapstructure:"JWT_ACCESS_TTL_MINUTES"`
		RefreshTTLDays int    `mapstructure:"JWT_REFRESH_TTL_DAYS"`
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

	// Uploads defaults
	v.SetDefault("UPLOADS_BASE_DIR", "./uploads")
	v.SetDefault("UPLOADS_BASE_URL", "/uploads")
	v.SetDefault("UPLOADS_MAX_MB", 5)

	// JWT defaults (development-safe but should be overridden in production)
	v.SetDefault("JWT_ACCESS_SECRET", "dev_access_secret_change_me")
	v.SetDefault("JWT_REFRESH_SECRET", "dev_refresh_secret_change_me")
	v.SetDefault("JWT_ACCESS_TTL_MINUTES", 15)
	v.SetDefault("JWT_REFRESH_TTL_DAYS", 7)

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
