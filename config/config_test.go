package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
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
ACCESS_TOKEN_SECRET=test_access_secret
ACCESS_TOKEN_LIFETIME_MINUTES=15
REFRESH_TOKEN_SECRET=test_refresh_secret
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
		s.Equal("test_access_secret", cfg.AccessTokenSecret)
		s.Equal(15, cfg.AccessTokenLifetimeMinutes)
		s.Equal("test_refresh_secret", cfg.RefreshTokenSecret)
		s.Equal(1440, cfg.RefreshTokenLifetimeMinutes)
		s.Equal("test_issuer", cfg.TokenIssuer)
		s.Equal("test_audience", cfg.TokenAudience)
		s.Equal("Test Sender", cfg.MailerSenderName)
		s.Equal("test@example.com", cfg.MailerSenderEmail)
		s.Equal("https://example.com/recover", cfg.PasswordRecoveryBaseURL)
	})
}
