package cert

import (
	"context"
	"fmt"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a signer certificate from IBM Verify by label",
	Long:  `Fetches and displays the signer certificate with the given --label.`,
	RunE:  runGetCert,
}

var (
	getCertTenant       string
	getCertClientID     string
	getCertClientSecret string
	getCertLabel        string
)

func init() {
	CertCmd.AddCommand(getCmd)

	getCmd.Flags().StringVar(&getCertTenant, "tenant", "", "IBM Verify tenant URL (required)")
	getCmd.Flags().StringVar(&getCertClientID, "client-id", "", "Cert-manager client ID (required)")
	getCmd.Flags().StringVar(&getCertClientSecret, "client-secret", "", "Cert-manager client secret (required)")
	getCmd.Flags().StringVar(&getCertLabel, "label", "", "Signer certificate label (required)")

	_ = getCmd.MarkFlagRequired("tenant")
	_ = getCmd.MarkFlagRequired("client-id")
	_ = getCmd.MarkFlagRequired("client-secret")
	_ = getCmd.MarkFlagRequired("label")
}

func runGetCert(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	c, err := client.New(getCertTenant, client.WithClientCredentials(getCertClientID, getCertClientSecret))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	cert, err := c.Certs.Get(ctx, getCertLabel)
	if err != nil {
		return fmt.Errorf("get signer cert: %w", err)
	}

	if cert == nil {
		fmt.Printf("No certificate found with label %q\n", getCertLabel)
		return nil
	}

	fmt.Printf("Label:   %s\n", cert.Label)
	fmt.Printf("Subject: %s\n", cert.Subject)
	fmt.Printf("Issuer:  %s\n", cert.Issuer)
	fmt.Printf("Cert:\n%s\n", cert.Cert)
	return nil
}
