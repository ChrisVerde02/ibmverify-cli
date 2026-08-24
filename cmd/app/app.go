// Package app contains the "ibmverify app" subcommands.
package app

import (
	"github.com/spf13/cobra"
)

// AppCmd is the "ibmverify app" parent command.
var AppCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage IBM Verify applications",
}
