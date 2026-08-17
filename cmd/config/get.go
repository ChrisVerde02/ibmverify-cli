package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Read a value from ~/.ibmverify/config.yaml",
	Long: `Read one or all values from your config file.

Valid keys:
  tenant, sts-client-id, sts-client-secret,
  cert-manager-client-id, cert-manager-client-secret,
  subject, issuer, label

Examples:
  ibmverify config get tenant
  ibmverify config get          # prints all keys`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGet,
}

func init() {
	ConfigCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	path := filepath.Join(home, ".ibmverify", "config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file not found — run 'ibmverify config init' first")
		}
		return fmt.Errorf("read config file: %w", err)
	}

	current := map[string]string{}
	if err := yaml.Unmarshal(data, &current); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}

	// Print a single key
	if len(args) == 1 {
		key := args[0]
		val, ok := current[key]
		if !ok {
			return fmt.Errorf("unknown config key %q — run 'ibmverify config get' to see all keys", key)
		}
		fmt.Fprintln(cmd.OutOrStdout(), val)
		return nil
	}

	// Print all keys in a consistent order
	keys := []string{
		"tenant",
		"sts-client-id",
		"sts-client-secret",
		"cert-manager-client-id",
		"cert-manager-client-secret",
		"subject",
		"issuer",
		"label",
	}
	for _, k := range keys {
		fmt.Fprintf(cmd.OutOrStdout(), "%-30s %s\n", k, current[k])
	}
	return nil
}
