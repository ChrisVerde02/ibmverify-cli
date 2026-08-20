// Package cmd contains all ibmverify CLI commands.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ChrisVerde02/ibmverify-cli/internal/errkind"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
)

// GlobalOutput holds the value of the --output flag, readable by all commands.
var GlobalOutput output.Format = output.Text

// GlobalDebug holds the value of the --debug flag.
var GlobalDebug bool

var rootCmd = &cobra.Command{
	Use:   "ibmverify",
	Short: "IBM Verify CLI — manage tokens and certificates from the terminal",
	Long: `ibmverify is a command-line tool for IBM Verify.

Use it to get access tokens, introspect tokens, manage signer certificates,
or run the full token-exchange flow in a single command.

Environment variables (override with flags):
  VERIFY_TENANT                IBM Verify tenant URL
  VERIFY_STS_CLIENT_ID         STS OAuth client ID
  VERIFY_STS_CLIENT_SECRET     STS OAuth client secret
  VERIFY_CERT_CLIENT_ID        Cert-manager client ID
  VERIFY_CERT_CLIENT_SECRET    Cert-manager client secret
  VERIFY_SUBJECT               JWT subject (IBM Verify username)
  VERIFY_ISSUER                JWT issuer claim
  VERIFY_LABEL                 Signer cert label / JWT kid

Config file: ~/.ibmverify/config.yaml (same keys, without VERIFY_ prefix).`,
}

// Root returns the root command so that main.go can attach subcommand packages.
func Root() *cobra.Command {
	return rootCmd
}

// Execute runs the root command. version is injected from main via ldflags.
func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(errkind.ExitCode(err))
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(
		(*string)(&GlobalOutput), "output", "o", string(output.Text),
		"Output format: text, json, yaml",
	)
	rootCmd.PersistentFlags().BoolVar(
		&GlobalDebug, "debug", false,
		"Print verbose HTTP debug info to stderr (secrets redacted)",
	)

	// Bind env vars through Viper for every common flag
	_ = viper.BindEnv("tenant", "VERIFY_TENANT")
	_ = viper.BindEnv("sts-client-id", "VERIFY_STS_CLIENT_ID")
	_ = viper.BindEnv("sts-client-secret", "VERIFY_STS_CLIENT_SECRET")
	_ = viper.BindEnv("cert-manager-client-id", "VERIFY_CERT_CLIENT_ID")
	_ = viper.BindEnv("cert-manager-client-secret", "VERIFY_CERT_CLIENT_SECRET")
	_ = viper.BindEnv("subject", "VERIFY_SUBJECT")
	_ = viper.BindEnv("issuer", "VERIFY_ISSUER")
	_ = viper.BindEnv("label", "VERIFY_LABEL")
}

// initConfig loads .env then ~/.ibmverify/config.yaml into Viper.
func initConfig() {
	// Load .env file if present (never fatal if missing)
	_ = godotenv.Load()

	viper.SetEnvPrefix("VERIFY")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Config file: ~/.ibmverify/config.yaml
	home, err := os.UserHomeDir()
	if err == nil {
		viper.AddConfigPath(fmt.Sprintf("%s/.ibmverify", home))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		_ = viper.ReadInConfig() // not fatal if absent
	}
}
