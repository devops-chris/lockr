package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/devops-chris/lockr/internal/ssm"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var readQuiet bool

var readCmd = &cobra.Command{
	Use:   "read [path]",
	Short: "Read a secret from SSM Parameter Store",
	Long: `Read a secret from AWS SSM Parameter Store.

Without a path, opens interactive search to find and read a secret.

Examples:
  # Interactive search, then read
  lockr read

  # Read a specific secret
  lockr read /myapp/prod/api-key

  # Output as JSON
  lockr read /myapp/prod/api-key --output json

  # Quiet mode (value only, for scripts)
  lockr read /myapp/prod/api-key --quiet`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRead,
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().BoolVarP(&readQuiet, "quiet", "q", false, "output value only (for scripts)")
}

func runRead(cmd *cobra.Command, args []string) error {
	var path string

	// If no path provided, do interactive search first
	if len(args) == 0 {
		selectedPath, err := interactiveSecretSearch()
		if err != nil {
			return err
		}
		if selectedPath == "" {
			return nil // User cancelled
		}
		path = selectedPath
	} else {
		path = buildPath(args[0])
	}

	client, err := ssm.NewClient(cfg.Region)
	if err != nil {
		return fmt.Errorf("failed to create SSM client: %w", err)
	}

	secret, err := client.ReadSecret(path)
	if err != nil {
		pterm.Error.Println("Failed to read secret")
		return fmt.Errorf("failed to read secret: %w", err)
	}

	// Quiet mode - just output the value
	if readQuiet {
		fmt.Print(secret.Value)
		return nil
	}

	switch cfg.Output {
	case "json":
		output := map[string]interface{}{
			"name":    secret.Name,
			"value":   secret.Value,
			"type":    secret.Type,
			"version": secret.Version,
		}
		if len(secret.Tags) > 0 {
			output["tags"] = secret.Tags
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	default:
		// Pretty output
		fmt.Println()
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
			WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
			Println("Secret")

		tableData := pterm.TableData{
			{"Property", "Value"},
			{"Name", pterm.FgCyan.Sprint(secret.Name)},
			{"Value", pterm.FgGreen.Sprint(secret.Value)},
			{"Type", secret.Type},
			{"Version", fmt.Sprintf("%d", secret.Version)},
		}

		pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()

		if len(secret.Tags) > 0 {
			fmt.Println()
			pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
				WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
				Println("Tags")

			tagData := pterm.TableData{{"Key", "Value"}}
			for k, v := range secret.Tags {
				tagData = append(tagData, []string{k, v})
			}
			pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tagData).Render()
		}
		fmt.Println()
	}

	return nil
}

// interactiveSecretSearch fetches all secrets and lets user search/select
func interactiveSecretSearch() (string, error) {
	spinner, _ := pterm.DefaultSpinner.Start("Fetching secrets...")

	client, err := ssm.NewClient(cfg.Region)
	if err != nil {
		spinner.Fail("Failed to create SSM client")
		return "", fmt.Errorf("failed to create SSM client: %w", err)
	}

	secrets, err := client.ListSecrets("/", true)
	if err != nil {
		spinner.Fail("Failed to list secrets")
		return "", fmt.Errorf("failed to list secrets: %w", err)
	}

	spinner.Stop()

	if len(secrets) == 0 {
		pterm.Warning.Println("No secrets found")
		return "", nil
	}

	// Build options
	options := make([]string, len(secrets))
	for i, s := range secrets {
		options[i] = s.Name
	}

	fmt.Println()
	pterm.Info.Printf("Found %d secrets\n", len(secrets))
	pterm.FgGray.Println("Type to filter • Enter to select • Ctrl+C to cancel")
	fmt.Println()

	selected, err := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithFilter(true).
		WithMaxHeight(20).
		Show()

	if err != nil {
		return "", nil // User cancelled
	}

	return selected, nil
}
