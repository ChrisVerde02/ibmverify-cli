// Package config contains the "ibmverify config" subcommands.
package config

import (
	"github.com/spf13/cobra"
)

// ConfigCmd is the "ibmverify config" parent command.
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ibmverify CLI configuration",
}
