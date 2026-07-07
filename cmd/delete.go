package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/devops-chris/clihq/ui"
	"github.com/devops-chris/lockr/internal/ssm"
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
		fmt.Println(ui.Warningf("You are about to delete: %s", ui.Error(path)))
		fmt.Println()

		var confirmed bool
		confirm := huh.NewConfirm().
			Title("Are you sure you want to delete this secret?").
			Value(&confirmed)
		confirm.WithTheme(ui.Theme())
		if err := confirm.Run(); err != nil {
			return err
		}

		if !confirmed {
			fmt.Println(ui.Info("Cancelled"))
			return nil
		}
	}

	client, err := ssm.NewClient(cfg.Region)
	if err != nil {
		return fmt.Errorf("failed to create SSM client: %w", err)
	}

	var deleteErr error
	_ = spinner.New().
		Title("Deleting secret...").
		Action(func() {
			deleteErr = client.DeleteSecret(path)
		}).
		Run()

	if deleteErr != nil {
		fmt.Println(ui.Error("Failed to delete secret"))
		return fmt.Errorf("failed to delete secret: %w", deleteErr)
	}

	fmt.Println()
	fmt.Println(ui.Successf("Deleted: %s", path))
	fmt.Println()

	return nil
}
