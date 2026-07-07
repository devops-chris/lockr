package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh/spinner"
	"github.com/devops-chris/clihq/ui"
	"github.com/devops-chris/lockr/internal/ssm"
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
		fmt.Println(ui.Error("Failed to read secret"))
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
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	default:
		fmt.Println()
		fmt.Println(ui.SectionHeader("Secret"))
		fmt.Println()

		rows := [][]string{
			{"Name", secret.Name},
			{"Value", ui.Highlight(secret.Value)},
			{"Type", secret.Type},
			{"Version", fmt.Sprintf("%d", secret.Version)},
		}
		fmt.Println(ui.Table([]string{"Property", "Value"}, rows))

		if len(secret.Tags) > 0 {
			fmt.Println()
			fmt.Println(ui.SectionHeader("Tags"))
			fmt.Println()

			tagRows := make([][]string, 0, len(secret.Tags))
			for k, v := range secret.Tags {
				tagRows = append(tagRows, []string{k, v})
			}
			fmt.Println(ui.Table([]string{"Key", "Value"}, tagRows))
		}
		fmt.Println()
	}

	return nil
}

// interactiveSecretSearch fetches all secrets and lets user fuzzy-search/select
func interactiveSecretSearch() (string, error) {
	client, err := ssm.NewClient(cfg.Region)
	if err != nil {
		return "", fmt.Errorf("failed to create SSM client: %w", err)
	}

	var secrets []ssm.SecretMetadata
	var listErr error
	_ = spinner.New().
		Title("Fetching secrets...").
		Action(func() {
			secrets, listErr = client.ListSecrets("/", true)
		}).
		Run()

	if listErr != nil {
		fmt.Println(ui.Error("Failed to list secrets"))
		return "", fmt.Errorf("failed to list secrets: %w", listErr)
	}

	if len(secrets) == 0 {
		fmt.Println(ui.Warning("No secrets found"))
		return "", nil
	}

	items := make([]pickItem, len(secrets))
	for i, s := range secrets {
		items[i] = pickItem{display: s.Name, search: s.Name, value: s.Name}
	}

	fmt.Println()
	fmt.Println(ui.Infof("Found %d secrets", len(secrets)))
	fmt.Println(ui.Subtle("Type to filter • ↑/↓ to move • Enter to select • Esc to cancel"))
	fmt.Println()

	selected, ok := runPicker(items, 20)
	if !ok {
		return "", nil
	}

	return selected, nil
}
