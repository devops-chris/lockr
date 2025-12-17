package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/devops-chris/lockr/internal/ssm"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	writeValue     string
	writeFile      string
	writeTags      []string
	writeOverwrite bool
)

var writeCmd = &cobra.Command{
	Use:   "write <path>",
	Short: "Write a secret to SSM Parameter Store",
	Long: `Write a secret to AWS SSM Parameter Store.

If no value is provided, you'll be prompted to enter it securely.
The value will not appear in your shell history.

Examples:
  # Interactive (secure prompt)
  lockr write /myapp/prod/db-password

  # With value flag (use carefully - may appear in history)
  lockr write /myapp/prod/api-key --value "sk_live_xxx"

  # From file (great for certs, keys, JSON)
  lockr write /myapp/prod/tls-cert --file ./cert.pem

  # From stdin (for piping)
  cat cert.pem | lockr write /myapp/prod/tls-cert --value -
  echo "myvalue" | lockr write /myapp/prod/key --value -

  # With tags
  lockr write /myapp/prod/api-key --tag owner=platform --tag env=prod

  # With prefix and env configured
  export LOCKR_PREFIX=/infra/saas
  export LOCKR_ENV=prod
  lockr write stripe/secret-key
  # Creates: /infra/saas/prod/stripe/secret-key`,
	Args: cobra.ExactArgs(1),
	RunE: runWrite,
}

func init() {
	rootCmd.AddCommand(writeCmd)

	writeCmd.Flags().StringVarP(&writeValue, "value", "v", "", "secret value (use '-' to read from stdin)")
	writeCmd.Flags().StringVarP(&writeFile, "file", "f", "", "read secret value from file")
	writeCmd.Flags().StringSliceVarP(&writeTags, "tag", "t", nil, "tags in key=value format (can be repeated)")
	writeCmd.Flags().BoolVar(&writeOverwrite, "overwrite", true, "overwrite existing secret")
}

func runWrite(cmd *cobra.Command, args []string) error {
	path := buildPath(args[0])
	var value string

	// Determine value source: file > value flag > stdin prompt
	switch {
	case writeFile != "":
		// Read from file
		data, err := os.ReadFile(writeFile)
		if err != nil {
			pterm.Error.Printf("Failed to read file: %s\n", writeFile)
			return fmt.Errorf("failed to read file: %w", err)
		}
		value = string(data)

	case writeValue == "-":
		// Read from stdin (for piping)
		data, err := readStdin()
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		value = data

	case writeValue != "":
		// Use provided value
		value = writeValue

	default:
		// Interactive prompt
		var err error
		value, err = promptSecureValue("Enter secret value: ")
		if err != nil {
			return fmt.Errorf("failed to read value: %w", err)
		}
	}

	if value == "" {
		pterm.Error.Println("Value cannot be empty")
		return fmt.Errorf("value cannot be empty")
	}

	// Parse tags
	tags := make(map[string]string)
	for _, tag := range writeTags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid tag format: %s (expected key=value)", tag)
		}
		tags[parts[0]] = parts[1]
	}

	// Show spinner while writing
	spinner, _ := pterm.DefaultSpinner.Start("Writing secret...")

	// Create SSM client and write
	client, err := ssm.NewClient(cfg.Region)
	if err != nil {
		spinner.Fail("Failed to create SSM client")
		return fmt.Errorf("failed to create SSM client: %w", err)
	}

	err = client.WriteSecret(path, value, tags, writeOverwrite, cfg.KMSKey)
	if err != nil {
		spinner.Fail("Failed to write secret")
		return fmt.Errorf("failed to write secret: %w", err)
	}

	spinner.Success("Secret written successfully")

	// Success output
	fmt.Println()
	pterm.DefaultBox.WithTitle("Created").Println(path)

	if len(tags) > 0 {
		fmt.Println()
		pterm.FgGray.Println("Tags:")
		for k, v := range tags {
			pterm.Println("  " + k + ": " + v)
		}
	}

	fmt.Println()

	return nil
}

func buildPath(input string) string {
	// If input already starts with /, use as-is
	if strings.HasPrefix(input, "/") {
		return input
	}

	var parts []string

	// Add prefix if configured
	if cfg.Prefix != "" {
		prefix := strings.Trim(cfg.Prefix, "/")
		parts = append(parts, prefix)
	}

	// Add env if configured
	if cfg.Env != "" {
		parts = append(parts, cfg.Env)
	}

	// Add the input path
	parts = append(parts, input)

	return "/" + strings.Join(parts, "/")
}

func promptSecureValue(prompt string) (string, error) {
	pterm.Print(pterm.FgYellow.Sprint("? ") + prompt)

	// Check if we're reading from a terminal
	if term.IsTerminal(int(os.Stdin.Fd())) {
		// Secure input (no echo)
		value, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println() // newline after hidden input
		return string(value), err
	}

	// Reading from pipe/redirect
	return readStdin()
}

func readStdin() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if len(lines) > 0 || line != "" {
				lines = append(lines, line)
			}
			break
		}
		lines = append(lines, line)
	}
	result := strings.Join(lines, "")
	// Trim trailing newline only (preserve internal newlines for certs, etc.)
	return strings.TrimSuffix(result, "\n"), nil
}
