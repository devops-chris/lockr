package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration options
// Precedence: CLI flags > ENV vars > Config file > Defaults
type Config struct {
	// Prefix is prepended to paths that don't start with /
	// ENV: LOCKR_PREFIX
	Prefix string `mapstructure:"prefix"`

	// Env is the environment (prod, staging, dev) added to the path
	// ENV: LOCKR_ENV
	Env string `mapstructure:"env"`

	// Output format: text, json
	// ENV: LOCKR_OUTPUT
	Output string `mapstructure:"output"`

	// KMSKey is the KMS key alias for encryption
	// ENV: LOCKR_KMS_KEY
	// Default: alias/aws/ssm (AWS managed key)
	KMSKey string `mapstructure:"kms_key"`

	// Region overrides the AWS region
	// ENV: LOCKR_REGION (or AWS_REGION)
	Region string `mapstructure:"region"`
}

// DefaultConfig returns configuration with sane defaults
func DefaultConfig() *Config {
	return &Config{
		Prefix: "",
		Env:    "",
		Output: "text",
		KMSKey: "alias/aws/ssm", // AWS managed key - just works
		Region: "",             // Use AWS SDK default
	}
}

// Load reads configuration from file and environment
func Load(configFile string) *Config {
	cfg := DefaultConfig()

	v := viper.New()

	// Set defaults
	v.SetDefault("prefix", cfg.Prefix)
	v.SetDefault("env", cfg.Env)
	v.SetDefault("output", cfg.Output)
	v.SetDefault("kms_key", cfg.KMSKey)
	v.SetDefault("region", cfg.Region)

	// Environment variables
	v.SetEnvPrefix("LOCKR")
	v.AutomaticEnv()

	// Config file
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// Default config locations
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(filepath.Join(home, ".config", "lockr"))
			v.AddConfigPath(filepath.Join(home, ".lockr"))
		}
		v.AddConfigPath(".")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Read config file (ignore if not found)
	_ = v.ReadInConfig()

	// Unmarshal into struct
	_ = v.Unmarshal(cfg)

	return cfg
}
