package cert

import (
	"context"
	"fmt"
	"os"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/ChrisVerde02/ibmverify-cli/internal/auth"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a signer certificate to IBM Verify",
	Long: `Reads a PEM certificate from --cert-file and uploads it to IBM Verify
as a signer certificate with the given --label.`,
	RunE: runUpload,
}

var (
	uploadTenant       string
	uploadClientID     string
	uploadClientSecret string
	uploadCertFile     string
	uploadLabel        string
)

func init() {
	CertCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringVar(&uploadTenant, "tenant", "", "IBM Verify tenant URL (required)")
	uploadCmd.Flags().StringVar(&uploadClientID, "client-id", "", "Cert-manager client ID (required)")
	uploadCmd.Flags().StringVar(&uploadClientSecret, "client-secret", "", "Cert-manager client secret (required)")
	uploadCmd.Flags().StringVar(&uploadCertFile, "cert-file", "", "Path to PEM certificate file (required)")
	uploadCmd.Flags().StringVar(&uploadLabel, "label", "", "Signer certificate label (required)")

	_ = uploadCmd.MarkFlagRequired("tenant")
	_ = uploadCmd.MarkFlagRequired("client-id")
	_ = uploadCmd.MarkFlagRequired("client-secret")
	_ = uploadCmd.MarkFlagRequired("cert-file")
	_ = uploadCmd.MarkFlagRequired("label")
}

func runUpload(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	certPEM, err := os.ReadFile(uploadCertFile)
	if err != nil {
		return fmt.Errorf("read cert file: %w", err)
	}

	token, err := auth.GetClientCredentialsToken(ctx, uploadTenant, uploadClientID, uploadClientSecret)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	if err := client.ImportSignerCert(ctx, client.SignerCertRequest{
		TenantURL:      uploadTenant,
		AccessToken:    token,
		CertificatePEM: string(certPEM),
		Label:          uploadLabel,
	}); err != nil {
		return fmt.Errorf("upload signer cert: %w", err)
	}

	fmt.Printf("✓ Certificate uploaded (label=%s)\n", uploadLabel)
	return nil
}
