// Package apiclient contains the "ibmverify apiclient" subcommands.
package apiclient

import (
	"github.com/spf13/cobra"
)

// APIClientCmd is the "ibmverify apiclient" parent command.
var APIClientCmd = &cobra.Command{
	Use:   "apiclient",
	Short: "Manage IBM Verify API clients (Dynamic Client Registration)",
}
