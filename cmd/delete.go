package cmd

import (
	"fmt"

	"github.com/devops-chris/lockr/internal/ssm"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a secret from SSM Parameter Store",
	Long: `Delete a secret from AWS SSM Parameter Store.

By default, you'll be prompted to confirm deletion.
Use --force to skip confirmation.

Examples:
  # Delete with confirmation
  lockr delete /myapp/prod/old-key

  # Delete without confirmation
  lockr delete /myapp/prod/old-key --force`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
	path := buildPath(args[0])

	// Confirm deletion unless --force
	if !deleteForce {
		fmt.Println()
		pterm.Warning.Printf("You are about to delete: %s\n", pterm.FgRed.Sprint(path))
		fmt.Println()

		result, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultText("Are you sure you want to delete this secret?").
			WithDefaultValue(false).
			Show()

		if !result {
			pterm.Info.Println("Cancelled")
			return nil
		}
	}

	// Show spinner while deleting
	spinner, _ := pterm.DefaultSpinner.Start("Deleting secret...")

	client, err := ssm.NewClient(cfg.Region)
	if err != nil {
		spinner.Fail("Failed to create SSM client")
		return fmt.Errorf("failed to create SSM client: %w", err)
	}

	err = client.DeleteSecret(path)
	if err != nil {
		spinner.Fail("Failed to delete secret")
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	spinner.Success("Secret deleted")

	fmt.Println()
	pterm.Success.Printf("Deleted: %s\n", path)
	fmt.Println()

	return nil
}
