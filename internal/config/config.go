package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv         string          `mapstructure:"APP_ENV"`
	HTTPAddr       string          `mapstructure:"HTTP_ADDR"`
	DatabaseURL    string          `mapstructure:"DATABASE_URL"`
	JWT            JWTConfig       `mapstructure:"JWT"`
	PasswordHash   PasswordConfig  `mapstructure:"PASSWORD_HASH"`
	CORSOrigins    []string        `mapstructure:"CORS_ORIGINS"`
	FCM            FCMConfig       `mapstructure:"FCM"`
	Map            MapConfig       `mapstructure:"MAP"`
	Weather        WeatherConfig   `mapstructure:"WEATHER"`
	Storage        StorageConfig   `mapstructure:"STORAGE"`
	Payment        PaymentConfig   `mapstructure:"PAYMENT"`
	LogLevel       string          `mapstructure:"LOG_LEVEL"`
	RequestTimeout time.Duration   `mapstructure:"REQUEST_TIMEOUT"`
	Migration      MigrationConfig `mapstructure:"MIGRATION"`
}

type JWTConfig struct {
	SigningKey      string        `mapstructure:"SIGNING_KEY"`
	Issuer          string        `mapstructure:"ISSUER"`
	Audience        string        `mapstructure:"AUDIENCE"`
	AccessTokenTTL  time.Duration `mapstructure:"ACCESS_TOKEN_TTL"`
	RefreshTokenTTL time.Duration `mapstructure:"REFRESH_TOKEN_TTL"`
}

type PasswordConfig struct {
	Algorithm   string `mapstructure:"ALGORITHM"`
	Memory      uint32 `mapstructure:"MEMORY"`
	Iterations  uint32 `mapstructure:"ITERATIONS"`
	Parallelism uint8  `mapstructure:"PARALLELISM"`
	SaltLength  uint32 `mapstructure:"SALT_LENGTH"`
	KeyLength   uint32 `mapstructure:"KEY_LENGTH"`
}

type FCMConfig struct {
	CredentialsPath string `mapstructure:"CREDENTIALS_PATH"`
	ProjectID       string `mapstructure:"PROJECT_ID"`
}

type MapConfig struct {
	Provider          string `mapstructure:"PROVIDER"`
	TileURL           string `mapstructure:"TILE_URL"`
	RoutingProvider   string `mapstructure:"ROUTING_PROVIDER"`
	OSRMBaseURL       string `mapstructure:"OSRM_BASE_URL"`
	GeocodingProvider string `mapstructure:"GEOCODING_PROVIDER"`
	GeocodingBaseURL  string `mapstructure:"GEOCODING_BASE_URL"`
}

type WeatherConfig struct {
	APIKey string `mapstructure:"API_KEY"`
}

type StorageConfig struct {
	Endpoint        string `mapstructure:"ENDPOINT"`
	Bucket          string `mapstructure:"BUCKET"`
	AccessKeyID     string `mapstructure:"ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"SECRET_ACCESS_KEY"`
	Region          string `mapstructure:"REGION"`
	UseSSL          bool   `mapstructure:"USE_SSL"`
}

type PaymentConfig struct {
	RazorpayKeyID     string `mapstructure:"RAZORPAY_KEY_ID"`
	RazorpayKeySecret string `mapstructure:"RAZORPAY_KEY_SECRET"`
}

type MigrationConfig struct {
	Path string `mapstructure:"PATH"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/vendorsaathi")

	// Allow environment variables to override config file values.
	// Nested config keys use underscore as separator in env vars:
	//   JWT.SIGNING_KEY  → JWT_SIGNING_KEY
	//   JWT.ISSUER       → JWT_ISSUER
	//   MIGRATION.PATH   → MIGRATION_PATH
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicitly bind every env-var name we care about.  AutomaticEnv maps
	// the viper key to an uppercased version; BindEnv is needed for:
	//   • flat keys whose viper name exactly matches the env var (e.g. DATABASE_URL)
	//   • nested keys that must be addressable by a flat env var (e.g. JWT_SIGNING_KEY)
	_ = v.BindEnv("DATABASE_URL", "DATABASE_URL")
	_ = v.BindEnv("APP_ENV", "APP_ENV")
	_ = v.BindEnv("HTTP_ADDR", "HTTP_ADDR")
	_ = v.BindEnv("LOG_LEVEL", "LOG_LEVEL")
	_ = v.BindEnv("REQUEST_TIMEOUT", "REQUEST_TIMEOUT")
	_ = v.BindEnv("JWT.SIGNING_KEY", "JWT_SIGNING_KEY")
	_ = v.BindEnv("JWT.ISSUER", "JWT_ISSUER")
	_ = v.BindEnv("JWT.AUDIENCE", "JWT_AUDIENCE")
	_ = v.BindEnv("JWT.ACCESS_TOKEN_TTL", "JWT_ACCESS_TOKEN_TTL")
	_ = v.BindEnv("JWT.REFRESH_TOKEN_TTL", "JWT_REFRESH_TOKEN_TTL")
	_ = v.BindEnv("MIGRATION.PATH", "MIGRATION_PATH")
	_ = v.BindEnv("MAP.PROVIDER", "MAP_PROVIDER")
	_ = v.BindEnv("MAP.TILE_URL", "MAP_TILE_URL")
	_ = v.BindEnv("MAP.ROUTING_PROVIDER", "ROUTING_PROVIDER")
	_ = v.BindEnv("MAP.OSRM_BASE_URL", "OSRM_BASE_URL")
	_ = v.BindEnv("MAP.GEOCODING_PROVIDER", "GEOCODING_PROVIDER")
	_ = v.BindEnv("MAP.GEOCODING_BASE_URL", "GEOCODING_BASE_URL")
	_ = v.BindEnv("PAYMENT.RAZORPAY_KEY_ID", "RAZORPAY_KEY_ID")
	_ = v.BindEnv("PAYMENT.RAZORPAY_KEY_SECRET", "RAZORPAY_KEY_SECRET")

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// No config file — environment-only mode is fine for Docker.
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// CORS_ORIGINS may be supplied as a JSON array string from an environment
	// variable (e.g. '["http://localhost:3000","http://localhost:5173"]').
	// Viper does not parse JSON arrays automatically when the value comes from
	// an environment variable, so we do it here.
	if len(cfg.CORSOrigins) == 0 {
		if raw := os.Getenv("CORS_ORIGINS"); raw != "" {
			var origins []string
			if err := json.Unmarshal([]byte(raw), &origins); err == nil {
				cfg.CORSOrigins = origins
			} else {
				// Fallback: treat as a comma-separated list.
				for _, o := range strings.Split(raw, ",") {
					if trimmed := strings.TrimSpace(o); trimmed != "" {
						cfg.CORSOrigins = append(cfg.CORSOrigins, trimmed)
					}
				}
			}
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("HTTP_ADDR", ":8080")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("REQUEST_TIMEOUT", 30*time.Second)
	v.SetDefault("MIGRATION.PATH", "file://migrations")

	v.SetDefault("JWT.ISSUER", "vendorsaathi")
	v.SetDefault("JWT.AUDIENCE", "vendorsaathi-api")
	v.SetDefault("JWT.ACCESS_TOKEN_TTL", 15*time.Minute)
	v.SetDefault("JWT.REFRESH_TOKEN_TTL", 30*24*time.Hour)

	v.SetDefault("PASSWORD_HASH.ALGORITHM", "argon2id")
	v.SetDefault("PASSWORD_HASH.MEMORY", 64*1024)
	v.SetDefault("PASSWORD_HASH.ITERATIONS", 3)
	v.SetDefault("PASSWORD_HASH.PARALLELISM", 4)
	v.SetDefault("PASSWORD_HASH.SALT_LENGTH", 16)
	v.SetDefault("PASSWORD_HASH.KEY_LENGTH", 32)

	v.SetDefault("STORAGE.USE_SSL", true)

	v.SetDefault("MAP.PROVIDER", "maplibre")
	v.SetDefault("MAP.TILE_URL", "https://tile.openstreetmap.org/{z}/{x}/{y}.png")
	v.SetDefault("MAP.ROUTING_PROVIDER", "osrm")
	v.SetDefault("MAP.OSRM_BASE_URL", "https://router.project-osrm.org")
	v.SetDefault("MAP.GEOCODING_PROVIDER", "nominatim")
	v.SetDefault("MAP.GEOCODING_BASE_URL", "https://nominatim.openstreetmap.org")
}

func (c *Config) Validate() error {
	if c.AppEnv == "" {
		return fmt.Errorf("APP_ENV is required")
	}
	if c.HTTPAddr == "" {
		return fmt.Errorf("HTTP_ADDR is required")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required — set DATABASE_URL env var or database_url in config.yaml")
	}
	if c.JWT.SigningKey == "" {
		return fmt.Errorf("JWT signing key is required — set JWT_SIGNING_KEY env var or jwt.signing_key in config.yaml")
	}
	if c.JWT.AccessTokenTTL <= 0 {
		return fmt.Errorf("JWT_ACCESS_TOKEN_TTL must be a positive duration (e.g. 15m)")
	}
	if c.JWT.RefreshTokenTTL <= 0 {
		return fmt.Errorf("JWT_REFRESH_TOKEN_TTL must be a positive duration (e.g. 720h)")
	}
	if len(c.CORSOrigins) == 0 && c.AppEnv != "development" {
		return fmt.Errorf("CORS_ORIGINS must be set in non-development environments")
	}
	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}
