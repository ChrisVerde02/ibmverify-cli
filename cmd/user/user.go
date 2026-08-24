// Package user contains the "ibmverify user" subcommands.
package user

import (
	"github.com/spf13/cobra"
)

// UserCmd is the "ibmverify user" parent command.
var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage IBM Verify users (SCIM v2)",
}
