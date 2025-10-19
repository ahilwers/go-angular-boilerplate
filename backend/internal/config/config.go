package config

import (
	"fmt"
)

type Config struct {
	Service  ServiceConfig  `yaml:"service" mapstructure:"service"`
	Database DatabaseConfig `yaml:"database" mapstructure:"database"`
	Auth     AuthConfig     `yaml:"auth" mapstructure:"auth"`
	Logging  LoggingConfig  `yaml:"logging" mapstructure:"logging"`
	CORS     CORSConfig     `yaml:"cors" mapstructure:"cors"`
	Docs     DocsConfig     `yaml:"docs" mapstructure:"docs"`
}

type ServiceConfig struct {
	Host         string `yaml:"host" mapstructure:"host"`
	Port         int    `yaml:"port" mapstructure:"port"`
	ReadTimeout  int    `yaml:"read_timeout" mapstructure:"read_timeout"`   // in seconds
	WriteTimeout int    `yaml:"write_timeout" mapstructure:"write_timeout"` // in seconds
}

type DatabaseConfig struct {
	URI      string `yaml:"uri" mapstructure:"uri"`
	Database string `yaml:"database" mapstructure:"database"`
	Username string `yaml:"username,omitempty" mapstructure:"username"` // optional
	Password string `yaml:"password,omitempty" mapstructure:"password"` // optional
	Timeout  int    `yaml:"timeout" mapstructure:"timeout"`             // in seconds
}

type AuthConfig struct {
	Enabled      bool   `yaml:"enabled" mapstructure:"enabled"`
	Issuer       string `yaml:"issuer" mapstructure:"issuer"`
	ClientID     string `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret string `yaml:"client_secret" mapstructure:"client_secret"`
	JWKSURL      string `yaml:"jwks_url" mapstructure:"jwks_url"`
}

type LoggingConfig struct {
	Level      string      `yaml:"level" mapstructure:"level"`   // debug, info, warn, error
	Format     string      `yaml:"format" mapstructure:"format"` // console, json
	LokiConfig *LokiConfig `yaml:"loki,omitempty" mapstructure:"loki"`
}

type LokiConfig struct {
	URL         string `yaml:"url" mapstructure:"url"`
	BearerToken string `yaml:"bearer_token,omitempty" mapstructure:"bearer_token"`
}

type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins" mapstructure:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods" mapstructure:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers" mapstructure:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers" mapstructure:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" mapstructure:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" mapstructure:"max_age"`
}

type DocsConfig struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
}

func Load(configPath string) (*Config, error) {
	return LoadWithViper(configPath)
}

func (c *ServiceConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
