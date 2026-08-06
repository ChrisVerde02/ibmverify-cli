package token

import (
	"context"
	"fmt"
	"time"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/ChrisVerde02/ibmverify-go/crypto"
	"github.com/ChrisVerde02/ibmverify-cli/internal/auth"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an IBM Verify access token via JWT token exchange",
	Long: `Generates a self-signed certificate, signs a short-lived JWT with it,
uploads the certificate to IBM Verify as a signer cert, and exchanges
the JWT for an access token. Prints the access token to stdout.`,
	RunE: runGet,
}

var (
	getTenant                  string
	getSTSClientID             string
	getSTSClientSecret         string
	getCertManagerClientID     string
	getCertManagerClientSecret string
	getSubject                 string
	getIssuer                  string
	getLabel                   string
	getOrganization            string
	getCountry                 string
	getValidityDays            int
	getKeySize                 int
	getSubjectTokenType        string
)

func init() {
	TokenCmd.AddCommand(getCmd)

	getCmd.Flags().StringVar(&getTenant, "tenant", "", "IBM Verify tenant URL (required)")
	getCmd.Flags().StringVar(&getSTSClientID, "sts-client-id", "", "STS client ID (required)")
	getCmd.Flags().StringVar(&getSTSClientSecret, "sts-client-secret", "", "STS client secret (required)")
	getCmd.Flags().StringVar(&getCertManagerClientID, "cert-manager-client-id", "", "Cert-manager client ID (required)")
	getCmd.Flags().StringVar(&getCertManagerClientSecret, "cert-manager-client-secret", "", "Cert-manager client secret (required)")
	getCmd.Flags().StringVar(&getSubject, "subject", "", "JWT subject — IBM Verify username (required)")
	getCmd.Flags().StringVar(&getIssuer, "issuer", "", "JWT issuer claim (required)")
	getCmd.Flags().StringVar(&getLabel, "label", "", "Signer cert label / JWT kid (required)")
	getCmd.Flags().StringVar(&getOrganization, "organization", "IBM", "Certificate O field")
	getCmd.Flags().StringVar(&getCountry, "country", "US", "Certificate C field")
	getCmd.Flags().IntVar(&getValidityDays, "validity-days", 365, "Certificate validity in days")
	getCmd.Flags().IntVar(&getKeySize, "key-size", 4096, "RSA key size (2048, 3072, or 4096)")
	getCmd.Flags().StringVar(&getSubjectTokenType, "subject-token-type", "urn:demo:token-type:user-jwt", "Subject token type URN")

	_ = getCmd.MarkFlagRequired("tenant")
	_ = getCmd.MarkFlagRequired("sts-client-id")
	_ = getCmd.MarkFlagRequired("sts-client-secret")
	_ = getCmd.MarkFlagRequired("cert-manager-client-id")
	_ = getCmd.MarkFlagRequired("cert-manager-client-secret")
	_ = getCmd.MarkFlagRequired("subject")
	_ = getCmd.MarkFlagRequired("issuer")
	_ = getCmd.MarkFlagRequired("label")
}

func runGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cert, err := crypto.GenerateSelfSignedCertificate(crypto.CertificateRequest{
		CommonName:   getLabel,
		Organization: getOrganization,
		Country:      getCountry,
		ValidityDays: getValidityDays,
		KeySize:      getKeySize,
	})
	if err != nil {
		return fmt.Errorf("generate certificate: %w", err)
	}

	certToken, err := auth.GetClientCredentialsToken(ctx, getTenant, getCertManagerClientID, getCertManagerClientSecret)
	if err != nil {
		return fmt.Errorf("get cert-manager token: %w", err)
	}

	if err := client.ImportSignerCert(ctx, client.SignerCertRequest{
		TenantURL:      getTenant,
		AccessToken:    certToken,
		CertificatePEM: cert.CertificatePEM,
		Label:          getLabel,
	}); err != nil {
		return fmt.Errorf("upload signer cert: %w", err)
	}

	jwtID := fmt.Sprintf("%s-%d", getLabel, time.Now().UnixNano())
	jwt, err := crypto.GenerateSignedJWT(crypto.JWTRequest{
		Issuer:        getIssuer,
		Subject:       getSubject,
		KeyID:         getLabel,
		JWTID:         jwtID,
		PrivateKeyPEM: cert.PrivateKeyPEM,
		ExpiresIn:     15 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("sign JWT: %w", err)
	}

	exchanged, err := client.ExchangeToken(ctx, client.TokenExchangeRequest{
		TenantURL:        getTenant,
		ClientID:         getSTSClientID,
		ClientSecret:     getSTSClientSecret,
		SubjectToken:     jwt.Token,
		SubjectTokenType: getSubjectTokenType,
	})
	if err != nil {
		return fmt.Errorf("exchange token: %w", err)
	}

	// Print just the token — makes it pipeable: TOKEN=$(ibmverify token get ...)
	fmt.Println(exchanged.AccessToken)
	return nil
}
