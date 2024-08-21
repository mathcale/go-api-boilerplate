package config

import "github.com/spf13/viper"

type Config struct {
	LogLevel         string `mapstructure:"LOG_LEVEL"`
	WebServerPort    int    `mapstructure:"WEB_SERVER_PORT"`
	DatabaseHost     string `mapstructure:"DATABASE_HOST"`
	DatabasePort     int64  `mapstructure:"DATABASE_PORT"`
	DatabaseUser     string `mapstructure:"DATABASE_USER"`
	DatabasePassword string `mapstructure:"DATABASE_PASSWORD"`
	DatabaseName     string `mapstructure:"DATABASE_NAME"`
	DatabaseSSLMode  string `mapstructure:"DATABASE_SSL_MODE"`
}

func Load(path string) (*Config, error) {
	var c *Config

	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	return c, nil
}
