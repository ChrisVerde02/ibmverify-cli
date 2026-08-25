package user

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an IBM Verify user by ID",
	RunE:  runDeleteUser,
}

var (
	deleteTenant       string
	deleteClientID     string
	deleteClientSecret string
	deleteUserID       string
)

func init() {
	UserCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVar(&deleteTenant, "tenant", "", "IBM Verify tenant URL (required)")
	deleteCmd.Flags().StringVar(&deleteClientID, "client-id", "", "Client ID (required)")
	deleteCmd.Flags().StringVar(&deleteClientSecret, "client-secret", "", "Client secret (required)")
	deleteCmd.Flags().StringVar(&deleteUserID, "id", "", "User ID (required)")

	_ = deleteCmd.MarkFlagRequired("tenant")
	_ = deleteCmd.MarkFlagRequired("client-id")
	_ = deleteCmd.MarkFlagRequired("client-secret")
	_ = deleteCmd.MarkFlagRequired("id")
}

func runDeleteUser(c *cobra.Command, args []string) error {
	ctx := context.Background()

	cl, err := client.New(deleteTenant, client.WithClientCredentials(deleteClientID, deleteClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	if err := cl.Users.Delete(ctx, deleteUserID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	fmt.Fprintf(c.OutOrStdout(), "✓ User deleted (id=%s)\n", deleteUserID)
	return nil
}
