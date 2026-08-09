package cert

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a signer certificate from IBM Verify",
	Long:  `Deletes the signer certificate with the given --label from IBM Verify.`,
	RunE:  runDelete,
}

var (
	deleteTenant       string
	deleteClientID     string
	deleteClientSecret string
	deleteLabel        string
)

func init() {
	CertCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVar(&deleteTenant, "tenant", "", "IBM Verify tenant URL (required)")
	deleteCmd.Flags().StringVar(&deleteClientID, "client-id", "", "Cert-manager client ID (required)")
	deleteCmd.Flags().StringVar(&deleteClientSecret, "client-secret", "", "Cert-manager client secret (required)")
	deleteCmd.Flags().StringVar(&deleteLabel, "label", "", "Signer certificate label (required)")

	_ = deleteCmd.MarkFlagRequired("tenant")
	_ = deleteCmd.MarkFlagRequired("client-id")
	_ = deleteCmd.MarkFlagRequired("client-secret")
	_ = deleteCmd.MarkFlagRequired("label")
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	tokenResult, err := client.GetClientCredentialsToken(ctx, client.ClientCredentialsRequest{
		TenantURL:    deleteTenant,
		ClientID:     deleteClientID,
		ClientSecret: deleteClientSecret,
	})
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	if err := client.DeleteSignerCert(ctx, deleteTenant, deleteLabel, tokenResult.AccessToken); err != nil {
		return fmt.Errorf("delete signer cert: %w", err)
	}

	fmt.Printf("✓ Certificate deleted (label=%s)\n", deleteLabel)
	return nil
}
