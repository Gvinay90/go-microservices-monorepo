package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

type ServerConfig struct {
	Port      string `mapstructure:"port"`
	Name      string `mapstructure:"name"`
	RateLimit int    `mapstructure:"rate_limit"` // Requests per minute
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"` // Data Source Name (connection string or file path)
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json or text
}

type AuthConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	KeycloakURL  string `mapstructure:"keycloak_url"`
	Realm        string `mapstructure:"realm"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	AuditLogging bool   `mapstructure:"audit_logging"`
}

// Load loads configuration from a YAML file and environment variables
// path: directory containing config file
// name: name of the config file (without extension)
func Load(path string, name string) (*Config, error) {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType("yaml")
	v.AddConfigPath(path)

	// Environment variable overrides
	// APP_SERVER_PORT overrides server.port
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
