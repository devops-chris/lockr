package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/devops-chris/lockr/internal/ssm"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	listRecursive   bool
	listInteractive bool
)

var listCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List secrets in SSM Parameter Store",
	Long: `List secrets in AWS SSM Parameter Store.

Without a path, lists ALL secrets (that you have access to) with interactive fuzzy search.
With a path, lists secrets at that path.

Examples:
  # List ALL secrets with fuzzy search (interactive)
  lockr list

  # List secrets at a path
  lockr list /myapp/prod

  # List recursively
  lockr list /myapp --recursive

  # Force interactive mode on a path
  lockr list /myapp -i

  # Output as JSON
  lockr list /myapp/prod --output json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&listRecursive, "recursive", "r", false, "list recursively")
	listCmd.Flags().BoolVarP(&listInteractive, "interactive", "i", false, "enable interactive fuzzy search")
}

func runList(cmd *cobra.Command, args []string) error {
	// Default to root path if none provided
	path := "/"
	if len(args) > 0 {
		path = buildPath(args[0])
	}

	// If no path provided, default to recursive and interactive
	noPathProvided := len(args) == 0
	if noPathProvided {
		listRecursive = true
		listInteractive = true
	}

	// Show spinner while fetching
	spinner, _ := pterm.DefaultSpinner.Start("Fetching secrets...")

	client, err := ssm.NewClient(cfg.Region)
	if err != nil {
		spinner.Fail("Failed to create SSM client")
		return fmt.Errorf("failed to create SSM client: %w", err)
	}

	secrets, err := client.ListSecrets(path, listRecursive)
	if err != nil {
		spinner.Fail("Failed to list secrets")
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	_ = spinner.Stop()

	if len(secrets) == 0 {
		pterm.Warning.Println("No secrets found at " + path)
		return nil
	}

	switch cfg.Output {
	case "json":
		data, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	default:
		// Interactive fuzzy search mode
		if listInteractive {
			return runInteractiveList(secrets)
		}

		// Standard table output
		return runTableList(secrets, path)
	}

	return nil
}

func runInteractiveList(secrets []ssm.SecretMetadata) error {
	// Build options for fuzzy search
	options := make([]string, len(secrets))
	for i, s := range secrets {
		options[i] = s.Name
	}

	fmt.Println()
	pterm.Info.Printf("Found %d secrets\n", len(secrets))
	pterm.FgGray.Println("Type to filter • Enter to select • Ctrl+C to exit")
	fmt.Println()

	// Interactive fuzzy select
	selected, err := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithFilter(true).
		WithMaxHeight(20).
		Show()

	if err != nil {
		// User cancelled
		return nil
	}

	// Find the selected secret and show details
	for _, s := range secrets {
		if s.Name == selected {
			fmt.Println()
			showSecretDetails(s)
			break
		}
	}

	return nil
}

func showSecretDetails(s ssm.SecretMetadata) {
	pterm.DefaultBox.WithTitle("Selected").Println(s.Name)

	fmt.Println()
	pterm.FgGray.Println("Details:")
	pterm.Println("  Type:     " + s.Type)
	pterm.Println("  Version:  " + fmt.Sprintf("%d", s.Version))
	if s.LastModified != nil {
		pterm.Println("  Modified: " + s.LastModified.Local().Format("2006-01-02 15:04:05"))
	}

	fmt.Println()
	pterm.FgGray.Println("To read the value:")
	pterm.Println("  lockr read " + s.Name)
	fmt.Println()
}

func runTableList(secrets []ssm.SecretMetadata, basePath string) error {
	fmt.Println()

	title := "All Secrets"
	if basePath != "/" {
		title = fmt.Sprintf("Secrets at %s", basePath)
	}

	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		Println(title)

	// Build table data
	tableData := pterm.TableData{
		{"Name", "Type", "Version", "Last Modified"},
	}

	for _, s := range secrets {
		// Show relative path if it starts with the search path
		displayName := s.Name
		if basePath != "/" && strings.HasPrefix(s.Name, basePath) {
			displayName = strings.TrimPrefix(s.Name, basePath)
			if displayName == "" {
				displayName = s.Name
			} else if strings.HasPrefix(displayName, "/") {
				displayName = displayName[1:]
			}
		}

		lastMod := "-"
		if s.LastModified != nil {
			lastMod = timeAgo(*s.LastModified)
		}

		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(displayName),
			s.Type,
			fmt.Sprintf("%d", s.Version),
			lastMod,
		})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()

	fmt.Println()
	pterm.Info.Printf("Total: %d secret(s)\n", len(secrets))

	// Hint about interactive mode
	if !listInteractive && len(secrets) > 10 {
		pterm.FgGray.Println("\nTip: Use 'lockr list -i' for interactive fuzzy search")
	}
	fmt.Println()

	return nil
}

// timeAgo returns a human-readable time difference
func timeAgo(t time.Time) string {
	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 30*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Local().Format("Jan 2, 2006")
	}
}
