package apiclient

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an IBM Verify API client by ID",
	RunE:  runDeleteAPIClient,
}

var (
	deleteTenant       string
	deleteClientID     string
	deleteClientSecret string
	deleteAPIClientID  string
)

func init() {
	APIClientCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVar(&deleteTenant, "tenant", "", "IBM Verify tenant URL (required)")
	deleteCmd.Flags().StringVar(&deleteClientID, "client-id", "", "Client ID (required)")
	deleteCmd.Flags().StringVar(&deleteClientSecret, "client-secret", "", "Client secret (required)")
	deleteCmd.Flags().StringVar(&deleteAPIClientID, "id", "", "API client ID (required)")

	_ = deleteCmd.MarkFlagRequired("tenant")
	_ = deleteCmd.MarkFlagRequired("client-id")
	_ = deleteCmd.MarkFlagRequired("client-secret")
	_ = deleteCmd.MarkFlagRequired("id")
}

func runDeleteAPIClient(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(deleteTenant, client.WithClientCredentials(deleteClientID, deleteClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	if err := cl.APIClients.Delete(ctx, deleteAPIClientID); err != nil {
		return fmt.Errorf("delete api client: %w", err)
	}

	fmt.Fprintf(c.OutOrStdout(), "✓ API client deleted (id=%s)\n", deleteAPIClientID)
	return nil
}
