package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	EnvironmentProduction  = "production"
	EnvironmentDevelopment = "development"
	EnvironmentTest        = "test"
)

type Config struct {
	Environment                 string `mapstructure:"ENVIRONMENT"`
	LogLevel                    string `mapstructure:"LOG_LEVEL"`
	WebServerPort               int    `mapstructure:"WEB_SERVER_PORT"`
	DatabaseHost                string `mapstructure:"DATABASE_HOST"`
	DatabasePort                int    `mapstructure:"DATABASE_PORT"`
	DatabaseUser                string `mapstructure:"DATABASE_USER"`
	DatabasePassword            string `mapstructure:"DATABASE_PASSWORD"`
	DatabaseName                string `mapstructure:"DATABASE_NAME"`
	DatabaseSSLMode             string `mapstructure:"DATABASE_SSL_MODE"`
	DatabaseMaxOpenConns        int    `mapstructure:"DATABASE_MAX_OPEN_CONNS"`
	DatabaseMaxIdleConns        int    `mapstructure:"DATABASE_MAX_IDLE_CONNS"`
	DatabaseConnMaxLifetimeSecs int    `mapstructure:"DATABASE_CONN_MAX_LIFETIME_SECS"`
	DatabaseConnMaxIdleTimeSecs int    `mapstructure:"DATABASE_CONN_MAX_IDLE_TIME_SECS"`
	AccessTokenSecret           string `mapstructure:"ACCESS_TOKEN_SECRET"`
	AccessTokenLifetimeMinutes  int    `mapstructure:"ACCESS_TOKEN_LIFETIME_MINUTES"`
	RefreshTokenSecret          string `mapstructure:"REFRESH_TOKEN_SECRET"`
	RefreshTokenLifetimeMinutes int    `mapstructure:"REFRESH_TOKEN_LIFETIME_MINUTES"`
	TokenIssuer                 string `mapstructure:"TOKEN_ISSUER"`
	TokenAudience               string `mapstructure:"TOKEN_AUDIENCE"`
	MailerSenderName            string `mapstructure:"MAILER_SENDER_NAME"`
	MailerSenderEmail           string `mapstructure:"MAILER_SENDER_EMAIL"`
	AccountConfirmationBaseURL  string `mapstructure:"ACCOUNT_CONFIRMATION_BASE_URL"`
	PasswordRecoveryBaseURL     string `mapstructure:"PASSWORD_RECOVERY_BASE_URL"`
	CORSAllowedOrigins          string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	RateLimitRequestsPerMinute  int    `mapstructure:"RATE_LIMIT_REQUESTS_PER_MINUTE"`
	RateLimitBurst              int    `mapstructure:"RATE_LIMIT_BURST"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}

	viper.SetConfigFile(filepath.Join(path, ".env"))
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
