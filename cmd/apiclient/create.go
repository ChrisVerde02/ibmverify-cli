package apiclient

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/ChrisVerde02/ibmverify-go/generated"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new IBM Verify API client",
	RunE:  runCreateAPIClient,
}

var (
	createTenant       string
	createClientID     string
	createClientSecret string
	createName         string
	createEntitlements []string
	createEnabled      bool
)

func init() {
	APIClientCmd.AddCommand(createCmd)

	createCmd.Flags().StringVar(&createTenant, "tenant", "", "IBM Verify tenant URL (required)")
	createCmd.Flags().StringVar(&createClientID, "client-id", "", "Client ID (required)")
	createCmd.Flags().StringVar(&createClientSecret, "client-secret", "", "Client secret (required)")
	createCmd.Flags().StringVar(&createName, "name", "", "API client name (required)")
	createCmd.Flags().StringSliceVar(&createEntitlements, "entitlements", []string{}, "Comma-separated list of entitlements")
	createCmd.Flags().BoolVar(&createEnabled, "enabled", true, "Whether the API client is enabled")

	_ = createCmd.MarkFlagRequired("tenant")
	_ = createCmd.MarkFlagRequired("client-id")
	_ = createCmd.MarkFlagRequired("client-secret")
	_ = createCmd.MarkFlagRequired("name")
}

func runCreateAPIClient(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(createTenant, client.WithClientCredentials(createClientID, createClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	result, err := cl.APIClients.Create(ctx, &generated.APIClientConfigRequest{
		ClientName:   createName,
		Entitlements: createEntitlements,
		Enabled:      createEnabled,
	})
	if err != nil {
		return fmt.Errorf("create api client: %w", err)
	}

	switch cmd.GlobalOutput {
	case output.JSON, output.YAML:
		return output.Print(c.OutOrStdout(), cmd.GlobalOutput, result)
	default:
		id, _ := result["clientId"].(string)
		fmt.Fprintf(c.OutOrStdout(), "✓ API client created (name=%s, id=%s)\n", createName, id)
	}
	return nil
}
