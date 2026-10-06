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

func (s *ConfigTestSuite) TearDownTest() {
	os.RemoveAll(s.tempDir)
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

		config, err := Load(s.tempDir)
		if err != nil {
			s.T().Logf("Load error: %v", err)
		}

		s.NoError(err)
		s.NotNil(config)
		s.Equal("development", config.Environment)
		s.Equal("debug", config.LogLevel)
		s.Equal(8080, config.WebServerPort)
		s.Equal("localhost", config.DatabaseHost)
		s.Equal(5432, config.DatabasePort)
		s.Equal("testuser", config.DatabaseUser)
		s.Equal("testpass", config.DatabasePassword)
		s.Equal("testdb", config.DatabaseName)
		s.Equal("disable", config.DatabaseSSLMode)
		s.Equal(10, config.DatabaseMaxOpenConns)
		s.Equal(5, config.DatabaseMaxIdleConns)
		s.Equal(300, config.DatabaseConnMaxLifetimeSecs)
		s.Equal(60, config.DatabaseConnMaxIdleTimeSecs)
		s.Equal("test_access_secret", config.AccessTokenSecret)
		s.Equal(15, config.AccessTokenLifetimeMinutes)
		s.Equal("test_refresh_secret", config.RefreshTokenSecret)
		s.Equal(1440, config.RefreshTokenLifetimeMinutes)
		s.Equal("test_issuer", config.TokenIssuer)
		s.Equal("test_audience", config.TokenAudience)
		s.Equal("Test Sender", config.MailerSenderName)
		s.Equal("test@example.com", config.MailerSenderEmail)
		s.Equal("https://example.com/recover", config.PasswordRecoveryBaseURL)
	})
}
