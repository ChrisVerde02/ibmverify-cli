package apiclient

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an IBM Verify API client by ID",
	RunE:  runGetAPIClient,
}

var (
	getTenant       string
	getClientID     string
	getClientSecret string
	getAPIClientID  string
)

func init() {
	APIClientCmd.AddCommand(getCmd)

	getCmd.Flags().StringVar(&getTenant, "tenant", "", "IBM Verify tenant URL (required)")
	getCmd.Flags().StringVar(&getClientID, "client-id", "", "Client ID (required)")
	getCmd.Flags().StringVar(&getClientSecret, "client-secret", "", "Client secret (required)")
	getCmd.Flags().StringVar(&getAPIClientID, "id", "", "API client ID (required)")

	_ = getCmd.MarkFlagRequired("tenant")
	_ = getCmd.MarkFlagRequired("client-id")
	_ = getCmd.MarkFlagRequired("client-secret")
	_ = getCmd.MarkFlagRequired("id")
}

func runGetAPIClient(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(getTenant, client.WithClientCredentials(getClientID, getClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	ac, err := cl.APIClients.Get(ctx, getAPIClientID)
	if err != nil {
		return fmt.Errorf("get api client: %w", err)
	}

	switch cmd.GlobalOutput {
	case output.JSON, output.YAML:
		return output.Print(c.OutOrStdout(), cmd.GlobalOutput, ac)
	default:
		id, _ := ac["clientId"].(string)
		name, _ := ac["clientName"].(string)
		enabled := fmt.Sprintf("%v", ac["enabled"])
		fmt.Fprintf(c.OutOrStdout(), "ID:      %s\nName:    %s\nEnabled: %s\n", id, name, enabled)
	}
	return nil
}
