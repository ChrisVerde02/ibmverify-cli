package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a value in ~/.ibmverify/config.yaml",
	Long: `Set a single key in your config file.

Valid keys:
  tenant, sts-client-id, sts-client-secret,
  cert-manager-client-id, cert-manager-client-secret,
  subject, issuer, label

Example:
  ibmverify config set tenant https://example.verify.ibm.com
  ibmverify config set subject myusername`,
	Args: cobra.ExactArgs(2),
	RunE: runSet,
}

func init() {
	ConfigCmd.AddCommand(setCmd)
}

var validKeys = map[string]bool{
	"tenant":                     true,
	"sts-client-id":              true,
	"sts-client-secret":          true,
	"cert-manager-client-id":     true,
	"cert-manager-client-secret": true,
	"subject":                    true,
	"issuer":                     true,
	"label":                      true,
}

func runSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	if !validKeys[key] {
		return fmt.Errorf("unknown config key %q — run 'ibmverify config set --help' for valid keys", key)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	path := filepath.Join(home, ".ibmverify", "config.yaml")

	// Load existing config or start fresh
	current := map[string]string{}
	if data, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(data, &current)
	}

	current[key] = value

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(current); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Set %s in %s\n", key, path)
	return nil
}
