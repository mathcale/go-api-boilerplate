package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/mathcale/go-api-boilerplate/internal/tests/fixtures"
)

type ConfigTestSuite struct {
	suite.Suite
	tempDir string
}

func (s *ConfigTestSuite) SetupTest() {
	viper.Reset()

	tempDir, err := os.MkdirTemp("", "config_test")
	s.NoError(err)
	s.tempDir = tempDir
}

func (s *ConfigTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *ConfigTestSuite) TearDownTest() {
	os.RemoveAll(s.tempDir)
}

func (s *ConfigTestSuite) TearDownSubTest() {
	s.TearDownTest()
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) TestLoad() {
	s.Run("should load config from .env file", func() {
		envContent := `ENVIRONMENT=development
LOG_LEVEL=debug
WEB_SERVER_PORT=8080
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=testuser
DATABASE_PASSWORD=testpass
DATABASE_NAME=testdb
DATABASE_SSL_MODE=disable
DATABASE_MAX_OPEN_CONNS=10
DATABASE_MAX_IDLE_CONNS=5
DATABASE_CONN_MAX_LIFETIME_SECS=300
DATABASE_CONN_MAX_IDLE_TIME_SECS=60
ACCESS_TOKEN_SECRET=` + fixtures.ConfigAccessSecret + `
ACCESS_TOKEN_LIFETIME_MINUTES=15
REFRESH_TOKEN_SECRET=` + fixtures.ConfigRefreshSecret + `
REFRESH_TOKEN_LIFETIME_MINUTES=1440
TOKEN_ISSUER=test_issuer
TOKEN_AUDIENCE=test_audience
MAILER_SENDER_NAME=Test Sender
MAILER_SENDER_EMAIL=test@example.com
PASSWORD_RECOVERY_BASE_URL=https://example.com/recover`

		envPath := filepath.Join(s.tempDir, ".env")
		err := os.WriteFile(envPath, []byte(envContent), 0644)
		s.NoError(err)

		cfg, err := Load(s.tempDir)
		if err != nil {
			s.T().Logf("Load error: %v", err)
		}

		s.NoError(err)
		s.NotNil(cfg)
		s.Equal("development", cfg.Environment)
		s.Equal("debug", cfg.LogLevel)
		s.Equal(8080, cfg.WebServerPort)
		s.Equal("localhost", cfg.DatabaseHost)
		s.Equal(5432, cfg.DatabasePort)
		s.Equal("testuser", cfg.DatabaseUser)
		s.Equal("testpass", cfg.DatabasePassword)
		s.Equal("testdb", cfg.DatabaseName)
		s.Equal("disable", cfg.DatabaseSSLMode)
		s.Equal(10, cfg.DatabaseMaxOpenConns)
		s.Equal(5, cfg.DatabaseMaxIdleConns)
		s.Equal(300, cfg.DatabaseConnMaxLifetimeSecs)
		s.Equal(60, cfg.DatabaseConnMaxIdleTimeSecs)
		s.Equal(fixtures.ConfigAccessSecret, cfg.AccessTokenSecret)
		s.Equal(15, cfg.AccessTokenLifetimeMinutes)
		s.Equal(fixtures.ConfigRefreshSecret, cfg.RefreshTokenSecret)
		s.Equal(1440, cfg.RefreshTokenLifetimeMinutes)
		s.Equal("test_issuer", cfg.TokenIssuer)
		s.Equal("test_audience", cfg.TokenAudience)
		s.Equal("Test Sender", cfg.MailerSenderName)
		s.Equal("test@example.com", cfg.MailerSenderEmail)
		s.Equal("https://example.com/recover", cfg.PasswordRecoveryBaseURL)
	})
	s.Run("should return an error when the loaded config is invalid", func() {
		envPath := filepath.Join(s.tempDir, ".env")
		s.Require().NoError(os.WriteFile(envPath, []byte("ENVIRONMENT=development"), 0600))

		cfg, err := Load(s.tempDir)

		s.Nil(cfg)
		s.Error(err)
	})
}

func (s *ConfigTestSuite) validConfig() *Config {
	return &Config{
		Environment:                 EnvironmentProduction,
		AccessTokenSecret:           fixtures.ConfigAccessSecret,
		RefreshTokenSecret:          fixtures.ConfigRefreshSecret,
		AccessTokenLifetimeMinutes:  15,
		RefreshTokenLifetimeMinutes: 1440,
		TokenIssuer:                 "issuer",
		TokenAudience:               "audience",
		CORSAllowedOrigins:          "https://example.com",
	}
}

func (s *ConfigTestSuite) TestValidate() {
	s.Run("should accept a valid config", func() {
		s.NoError(s.validConfig().Validate())
	})

	s.Run("should accept every known environment", func() {
		for _, env := range []string{
			EnvironmentProduction, EnvironmentDevelopment, EnvironmentTest,
		} {
			cfg := s.validConfig()
			cfg.Environment = env
			s.NoError(cfg.Validate(), env)
		}
	})

	s.Run("should accept empty CORS origins outside production", func() {
		cfg := s.validConfig()
		cfg.Environment = EnvironmentDevelopment
		cfg.CORSAllowedOrigins = ""
		s.NoError(cfg.Validate())
	})

	cases := []struct {
		name     string
		mutate   func(c *Config)
		contains string
	}{
		{"unknown environment", func(c *Config) { c.Environment = "staging" }, "ENVIRONMENT"},
		{"empty environment", func(c *Config) { c.Environment = "" }, "ENVIRONMENT"},
		{
			"short access secret",
			func(c *Config) { c.AccessTokenSecret = fixtures.ConfigShortValue },
			"ACCESS_TOKEN_SECRET must be at least",
		},
		{
			"empty access secret",
			func(c *Config) { c.AccessTokenSecret = "" },
			"ACCESS_TOKEN_SECRET must be at least",
		},
		{
			"short refresh secret",
			func(c *Config) { c.RefreshTokenSecret = fixtures.ConfigShortValue },
			"REFRESH_TOKEN_SECRET must be at least",
		},
		{
			"default access secret",
			func(c *Config) { c.AccessTokenSecret = fixtures.ConfigPlaceholderAccess },
			"ACCESS_TOKEN_SECRET must not use a default",
		},
		{
			"default refresh secret",
			func(c *Config) { c.RefreshTokenSecret = fixtures.ConfigPlaceholderRefresh },
			"REFRESH_TOKEN_SECRET must not use a default",
		},
		{
			"identical secrets",
			func(c *Config) { c.RefreshTokenSecret = c.AccessTokenSecret },
			"must differ",
		},
		{"empty issuer", func(c *Config) { c.TokenIssuer = "" }, "TOKEN_ISSUER"},
		{"empty audience", func(c *Config) { c.TokenAudience = "" }, "TOKEN_AUDIENCE"},
		{
			"zero access lifetime",
			func(c *Config) { c.AccessTokenLifetimeMinutes = 0 },
			"ACCESS_TOKEN_LIFETIME_MINUTES",
		},
		{
			"negative refresh lifetime",
			func(c *Config) { c.RefreshTokenLifetimeMinutes = -1 },
			"REFRESH_TOKEN_LIFETIME_MINUTES",
		},
		{
			"production without CORS origins",
			func(c *Config) { c.CORSAllowedOrigins = "  " },
			"CORS_ALLOWED_ORIGINS",
		},
	}

	for _, tc := range cases {
		s.Run("should reject "+tc.name, func() {
			cfg := s.validConfig()
			tc.mutate(cfg)

			err := cfg.Validate()

			s.Require().Error(err)
			s.Contains(err.Error(), tc.contains)
		})
	}

	s.Run("should report every violation at once", func() {
		err := (&Config{}).Validate()

		s.Require().Error(err)
		s.Contains(err.Error(), "ENVIRONMENT")
		s.Contains(err.Error(), "ACCESS_TOKEN_SECRET")
		s.Contains(err.Error(), "TOKEN_ISSUER")
		s.Contains(err.Error(), "REFRESH_TOKEN_LIFETIME_MINUTES")
	})

	s.Run("should never include secret values in the error", func() {
		secrets := []string{
			fixtures.ConfigShortValue,
			fixtures.ConfigPlaceholderAccess,
			fixtures.ConfigAccessSecret[:32],
		}

		for _, secret := range secrets {
			cfg := s.validConfig()
			cfg.AccessTokenSecret = secret
			cfg.RefreshTokenSecret = secret

			err := cfg.Validate()

			s.Require().Error(err)
			s.NotContains(err.Error(), secret)
		}
	})
}
