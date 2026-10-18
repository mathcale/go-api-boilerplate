package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	EnvironmentProduction  = "production"
	EnvironmentDevelopment = "development"
	EnvironmentTest        = "test"
)

const (
	MailerProviderFake = "fake"
	MailerProviderSMTP = "smtp"

	SMTPTLSModeStartTLS = "starttls"
	SMTPTLSModeImplicit = "implicit"
)

const minSecretLength = 64

var defaultSecrets = map[string]struct{}{
	"change-me-access-secret":  {},
	"change-me-refresh-secret": {},
}

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
	MailerProvider              string `mapstructure:"MAILER_PROVIDER"`
	SMTPHost                    string `mapstructure:"SMTP_HOST"`
	SMTPPort                    int    `mapstructure:"SMTP_PORT"`
	SMTPUsername                string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword                string `mapstructure:"SMTP_PASSWORD"`
	SMTPTLSMode                 string `mapstructure:"SMTP_TLS_MODE"`
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
	viper.SetDefault("MAILER_PROVIDER", MailerProviderFake)
	viper.SetDefault("SMTP_PORT", 0)
	viper.SetDefault("SMTP_HOST", "")
	viper.SetDefault("SMTP_USERNAME", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("SMTP_TLS_MODE", SMTPTLSModeStartTLS)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	var errs []error

	switch c.Environment {
	case EnvironmentProduction, EnvironmentDevelopment, EnvironmentTest:
	default:
		errs = append(errs, fmt.Errorf(
			"ENVIRONMENT must be one of %q, %q or %q",
			EnvironmentProduction, EnvironmentDevelopment, EnvironmentTest,
		))
	}

	errs = append(errs, validateSecret("ACCESS_TOKEN_SECRET", c.AccessTokenSecret)...)
	errs = append(errs, validateSecret("REFRESH_TOKEN_SECRET", c.RefreshTokenSecret)...)

	if c.AccessTokenSecret != "" && c.AccessTokenSecret == c.RefreshTokenSecret {
		errs = append(errs, errors.New(
			"ACCESS_TOKEN_SECRET and REFRESH_TOKEN_SECRET must differ",
		))
	}

	if strings.TrimSpace(c.TokenIssuer) == "" {
		errs = append(errs, errors.New("TOKEN_ISSUER must not be empty"))
	}

	if strings.TrimSpace(c.TokenAudience) == "" {
		errs = append(errs, errors.New("TOKEN_AUDIENCE must not be empty"))
	}

	if c.AccessTokenLifetimeMinutes <= 0 {
		errs = append(errs, errors.New("ACCESS_TOKEN_LIFETIME_MINUTES must be greater than 0"))
	}

	if c.RefreshTokenLifetimeMinutes <= 0 {
		errs = append(errs, errors.New("REFRESH_TOKEN_LIFETIME_MINUTES must be greater than 0"))
	}

	if c.Environment == EnvironmentProduction && strings.TrimSpace(c.CORSAllowedOrigins) == "" {
		errs = append(errs, errors.New("CORS_ALLOWED_ORIGINS must not be empty in production"))
	}

	errs = append(errs, c.validateMailer()...)

	return errors.Join(errs...)
}

func (c *Config) validateMailer() []error {
	switch c.MailerProvider {
	case MailerProviderFake:
		if c.Environment == EnvironmentProduction {
			return []error{errors.New("MAILER_PROVIDER must not be \"fake\" in production")}
		}

		return nil
	case MailerProviderSMTP:
	default:
		return []error{fmt.Errorf(
			"MAILER_PROVIDER must be one of %q or %q", MailerProviderFake, MailerProviderSMTP,
		)}
	}

	var errs []error

	if strings.TrimSpace(c.SMTPHost) == "" {
		errs = append(errs, errors.New("SMTP_HOST must not be empty when MAILER_PROVIDER is smtp"))
	}

	if c.SMTPPort <= 0 {
		errs = append(errs, errors.New(
			"SMTP_PORT must be greater than 0 when MAILER_PROVIDER is smtp",
		))
	}

	switch c.SMTPTLSMode {
	case SMTPTLSModeStartTLS, SMTPTLSModeImplicit:
	default:
		errs = append(errs, fmt.Errorf(
			"SMTP_TLS_MODE must be one of %q or %q", SMTPTLSModeStartTLS, SMTPTLSModeImplicit,
		))
	}

	return errs
}

func validateSecret(name, value string) []error {
	var errs []error

	if len(value) < minSecretLength {
		errs = append(errs, fmt.Errorf("%s must be at least %d bytes long", name, minSecretLength))
	}

	if _, ok := defaultSecrets[value]; ok {
		errs = append(errs, fmt.Errorf("%s must not use a default placeholder value", name))
	}

	return errs
}
