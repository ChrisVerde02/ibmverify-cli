// Package cmd contains all ibmverify CLI commands.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command — "ibmverify".
var rootCmd = &cobra.Command{
	Use:   "ibmverify",
	Short: "IBM Verify CLI — manage tokens and certificates from the terminal",
	Long: `ibmverify is a command-line tool for IBM Verify.

Use it to get access tokens, introspect tokens, manage signer certificates,
or run the full token-exchange flow in a single command.`,
}

// Root returns the root command so that main.go can attach subcommand packages.
func Root() *cobra.Command {
	return rootCmd
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
