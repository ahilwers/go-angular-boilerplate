package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func LoadWithViper(configPath string) (*Config, error) {
	v := viper.New()
	setDefaults(v)
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("local")
		v.AddConfigPath("./config")     // From project root
		v.AddConfigPath("../config")    // From backend/
		v.AddConfigPath("../../config") // From backend/internal/
		v.AddConfigPath(".")            // Current directory
	}

	// Enable environment variable support
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file (ignore if not found, we'll use defaults + env vars)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; use defaults and env vars
	}

	// Unmarshal into Config struct
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Service defaults
	v.SetDefault("service.host", "localhost")
	v.SetDefault("service.port", 8080)
	v.SetDefault("service.read_timeout", 10)
	v.SetDefault("service.write_timeout", 10)

	// Database defaults
	v.SetDefault("database.uri", "mongodb://localhost:27017")
	v.SetDefault("database.database", "boilerplate")
	v.SetDefault("database.timeout", 10)

	// Auth defaults
	v.SetDefault("auth.enabled", false)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "console")

	// Rate limit defaults
	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.requests_per_second", 10)
	v.SetDefault("rate_limit.burst", 20)
}
