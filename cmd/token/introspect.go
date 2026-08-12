package token

import (
	"context"
	"fmt"
	"time"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var introspectCmd = &cobra.Command{
	Use:   "introspect",
	Short: "Introspect an IBM Verify access token",
	Long:  `Calls IBM Verify's /oauth2/introspect endpoint and prints token metadata.`,
	RunE:  runIntrospect,
}

var (
	introspectTenant       string
	introspectClientID     string
	introspectClientSecret string
	introspectToken        string
)

func init() {
	TokenCmd.AddCommand(introspectCmd)

	introspectCmd.Flags().StringVar(&introspectTenant, "tenant", "", "IBM Verify tenant URL (required)")
	introspectCmd.Flags().StringVar(&introspectClientID, "client-id", "", "OAuth client ID (required)")
	introspectCmd.Flags().StringVar(&introspectClientSecret, "client-secret", "", "OAuth client secret (required)")
	introspectCmd.Flags().StringVar(&introspectToken, "token", "", "Access token to introspect (required)")

	_ = introspectCmd.MarkFlagRequired("tenant")
	_ = introspectCmd.MarkFlagRequired("client-id")
	_ = introspectCmd.MarkFlagRequired("client-secret")
	_ = introspectCmd.MarkFlagRequired("token")
}

func runIntrospect(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	c, err := client.New(introspectTenant, client.WithClientCredentials(introspectClientID, introspectClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	info, err := c.Token.Introspect(ctx, introspectToken)
	if err != nil {
		return fmt.Errorf("introspect token: %w", err)
	}

	status := "INACTIVE"
	if info.Active {
		status = "ACTIVE"
	}

	fmt.Printf("Status:   %s\n", status)
	fmt.Printf("Subject:  %s\n", info.Subject)
	fmt.Printf("Username: %s\n", info.Username)
	if info.PreferredUsername != "" {
		fmt.Printf("Preferred username: %s\n", info.PreferredUsername)
	}
	if info.Name != "" {
		fmt.Printf("Name:     %s\n", info.Name)
	}
	fmt.Printf("Scope:    %s\n", info.Scope)
	fmt.Printf("Issuer:   %s\n", info.Issuer)
	if info.ExpiresAt > 0 {
		exp := time.Unix(info.ExpiresAt, 0).UTC()
		fmt.Printf("Expires:  %s\n", exp.Format(time.RFC3339))
	}

	return nil
}
