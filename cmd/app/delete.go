package app

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an IBM Verify application by ID",
	RunE:  runDeleteApp,
}

var (
	deleteTenant       string
	deleteClientID     string
	deleteClientSecret string
	deleteAppID        string
)

func init() {
	AppCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVar(&deleteTenant, "tenant", "", "IBM Verify tenant URL (required)")
	deleteCmd.Flags().StringVar(&deleteClientID, "client-id", "", "Client ID (required)")
	deleteCmd.Flags().StringVar(&deleteClientSecret, "client-secret", "", "Client secret (required)")
	deleteCmd.Flags().StringVar(&deleteAppID, "id", "", "Application ID (required)")

	_ = deleteCmd.MarkFlagRequired("tenant")
	_ = deleteCmd.MarkFlagRequired("client-id")
	_ = deleteCmd.MarkFlagRequired("client-secret")
	_ = deleteCmd.MarkFlagRequired("id")
}

func runDeleteApp(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(deleteTenant, client.WithClientCredentials(deleteClientID, deleteClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	if err := cl.Apps.Delete(ctx, deleteAppID); err != nil {
		return fmt.Errorf("delete app: %w", err)
	}

	fmt.Fprintf(c.OutOrStdout(), "✓ Application deleted (id=%s)\n", deleteAppID)
	return nil
}
