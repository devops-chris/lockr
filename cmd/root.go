package cmd

import (
	"fmt"
	"os"

	"github.com/devops-chris/lockr/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfg       *config.Config
	cfgFile   string
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// SetVersion sets the version info from build flags
func SetVersion(v, c, d string) {
	version = v
	commit = c
	buildDate = d
}

var rootCmd = &cobra.Command{
	Use:   "lockr",
	Short: "Manage secrets in AWS SSM Parameter Store",
	Long: `lockr - A simple CLI for managing secrets in AWS SSM Parameter Store.

Works with zero configuration using sane defaults.
Configuration precedence: CLI flags > ENV vars > Config file > Defaults

Environment variables:
  LOCKR_PREFIX   Path prefix for relative paths (e.g., /infra/saas)
  LOCKR_ENV      Environment to include in path (e.g., prod, staging)
  LOCKR_OUTPUT   Output format: text, json (default: text)
  LOCKR_KMS_KEY  KMS key alias (default: alias/aws/ssm)
  LOCKR_REGION   AWS region (default: from AWS config)

Examples:
  # Write a secret (prompts for value)
  lockr write /myapp/prod/api-key

  # Write with prefix and env configured
  export LOCKR_PREFIX=/infra/saas
  export LOCKR_ENV=prod
  lockr write datadog/api-key
  # Creates: /infra/saas/prod/datadog/api-key

  # Read a secret
  lockr read /myapp/prod/api-key

  # List all secrets (interactive fuzzy search)
  lockr list

  # List secrets at a path
  lockr list /myapp/prod

  # Delete a secret
  lockr delete /myapp/prod/old-key`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/lockr/config.yaml)")
	rootCmd.PersistentFlags().String("prefix", "", "path prefix for secrets")
	rootCmd.PersistentFlags().String("env", "", "environment (e.g., prod, staging)")
	rootCmd.PersistentFlags().String("output", "text", "output format (text, json)")
	rootCmd.PersistentFlags().String("region", "", "AWS region (default: from AWS config)")
}

func initConfig() {
	cfg = config.Load(cfgFile)

	// Override with CLI flags if provided
	if prefix, _ := rootCmd.PersistentFlags().GetString("prefix"); prefix != "" {
		cfg.Prefix = prefix
	}
	if env, _ := rootCmd.PersistentFlags().GetString("env"); env != "" {
		cfg.Env = env
	}
	if output, _ := rootCmd.PersistentFlags().GetString("output"); output != "" {
		cfg.Output = output
	}
	if region, _ := rootCmd.PersistentFlags().GetString("region"); region != "" {
		cfg.Region = region
	}
}
