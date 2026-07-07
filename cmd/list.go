package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/devops-chris/clihq/ui"
	"github.com/devops-chris/lockr/internal/ssm"
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

	client, err := ssm.NewClient(cfg.Region)
	if err != nil {
		return fmt.Errorf("failed to create SSM client: %w", err)
	}

	var secrets []ssm.SecretMetadata
	var listErr error
	_ = spinner.New().
		Title("Fetching secrets...").
		Action(func() {
			secrets, listErr = client.ListSecrets(path, listRecursive)
		}).
		Run()

	if listErr != nil {
		fmt.Println(ui.Error("Failed to list secrets"))
		return fmt.Errorf("failed to list secrets: %w", listErr)
	}

	if len(secrets) == 0 {
		fmt.Println(ui.Warningf("No secrets found at %s", path))
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
		fmt.Println()
		fmt.Println(ui.Banner("lockr", "secrets manager for AWS SSM Parameter Store"))

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
	items := make([]pickItem, len(secrets))
	for i, s := range secrets {
		items[i] = pickItem{display: s.Name, search: s.Name, value: s.Name}
	}

	fmt.Println()
	fmt.Println(ui.Infof("Found %d secrets", len(secrets)))
	fmt.Println(ui.Subtle("Type to filter • ↑/↓ to move • Enter to select • Esc to exit"))
	fmt.Println()

	selected, ok := runPicker(items, 20)
	if !ok {
		return nil
	}

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
	fmt.Println(ui.SectionHeader("Selected"))
	fmt.Println(ui.Highlight(s.Name))
	fmt.Println()

	fmt.Println(ui.Subtle("Details:"))
	fmt.Println("  Type:     " + s.Type)
	fmt.Println("  Version:  " + fmt.Sprintf("%d", s.Version))
	if s.LastModified != nil {
		fmt.Println("  Modified: " + s.LastModified.Local().Format("2006-01-02 15:04:05"))
	}

	fmt.Println()
	fmt.Println(ui.Subtle("To read the value:"))
	fmt.Println("  lockr read " + s.Name)
	fmt.Println()
}

func runTableList(secrets []ssm.SecretMetadata, basePath string) error {
	fmt.Println()

	title := "All Secrets"
	if basePath != "/" {
		title = fmt.Sprintf("Secrets at %s", basePath)
	}
	fmt.Println(ui.SectionHeader(title))
	fmt.Println()

	headers := []string{"Name", "Type", "Version", "Last Modified"}
	rows := make([][]string, 0, len(secrets))

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

		rows = append(rows, []string{
			ui.Highlight(displayName),
			s.Type,
			fmt.Sprintf("%d", s.Version),
			lastMod,
		})
	}

	fmt.Println(ui.Table(headers, rows))

	fmt.Println()
	fmt.Println(ui.Infof("Total: %d secret(s)", len(secrets)))

	// Hint about interactive mode
	if !listInteractive && len(secrets) > 10 {
		fmt.Println()
		fmt.Println(ui.Subtle("Tip: Use 'lockr list -i' for interactive fuzzy search"))
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
