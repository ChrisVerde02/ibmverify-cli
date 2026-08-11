package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default config file at ~/.ibmverify/config.yaml",
	Long: `Creates ~/.ibmverify/config.yaml with empty placeholder values.

Edit the file to set your tenant, client IDs, and secrets so you don't
need to pass flags or environment variables on every command.

Precedence: flag > env var > config file > default.`,
	RunE: runInit,
}

func init() {
	ConfigCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	dir := filepath.Join(home, ".ibmverify")
	path := filepath.Join(dir, "config.yaml")

	// Don't overwrite an existing config
	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Config file already exists: %s\n", path)
		fmt.Fprintf(cmd.OutOrStdout(), "Use 'ibmverify config set' to update individual values.\n")
		return nil
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	defaults := map[string]string{
		"tenant":                   "",
		"sts-client-id":            "",
		"sts-client-secret":        "",
		"cert-manager-client-id":   "",
		"cert-manager-client-secret": "",
		"subject":                  "",
		"issuer":                   "",
		"label":                    "",
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(defaults); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Config file created: %s\n", path)
	fmt.Fprintf(cmd.OutOrStdout(), "  Edit it to set your credentials.\n")
	return nil
}
