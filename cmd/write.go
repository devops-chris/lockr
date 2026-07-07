package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/devops-chris/clihq/ui"
	"github.com/devops-chris/lockr/internal/ssm"
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
			fmt.Println(ui.Errorf("Failed to read file: %s", writeFile))
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
		value, err = promptSecureValue("Secret value")
		if err != nil {
			return fmt.Errorf("failed to read value: %w", err)
		}
	}

	if value == "" {
		fmt.Println(ui.Error("Value cannot be empty"))
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

	client, err := ssm.NewClient(cfg.Region)
	if err != nil {
		return fmt.Errorf("failed to create SSM client: %w", err)
	}

	var writeErr error
	_ = spinner.New().
		Title("Writing secret...").
		Action(func() {
			writeErr = client.WriteSecret(path, value, tags, writeOverwrite, cfg.KMSKey)
		}).
		Run()

	if writeErr != nil {
		fmt.Println(ui.Error("Failed to write secret"))
		return fmt.Errorf("failed to write secret: %w", writeErr)
	}

	fmt.Println(ui.Success("Secret written successfully"))
	fmt.Println()
	fmt.Println(ui.Subtle("Created: ") + ui.Highlight(path))

	if len(tags) > 0 {
		fmt.Println()
		fmt.Println(ui.Subtle("Tags:"))
		for k, v := range tags {
			fmt.Println("  " + k + ": " + v)
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

func promptSecureValue(title string) (string, error) {
	// Reading from a pipe/redirect - huh needs a real TTY, fall back to stdin
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return readStdin()
	}

	var value string
	input := huh.NewInput().
		Title(title).
		EchoMode(huh.EchoModePassword).
		Value(&value)
	input.WithTheme(ui.Theme())

	if err := input.Run(); err != nil {
		return "", err
	}

	return value, nil
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
