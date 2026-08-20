package token

import (
	"context"
	"fmt"
	"time"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/ChrisVerde02/ibmverify-go/crypto"
	"github.com/spf13/cobra"

	"github.com/ChrisVerde02/ibmverify-cli/internal/retry"
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
	getJWTExpiresIn            time.Duration
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
	getCmd.Flags().DurationVar(&getJWTExpiresIn, "jwt-expires-in", 15*time.Minute, "JWT lifetime (e.g. 15m, 1h). Controls how long the signed JWT is valid before exchange.")
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

	// cert-manager client — auto-obtains its own token internally
	certClient, err := client.New(getTenant, client.WithClientCredentials(getCertManagerClientID, getCertManagerClientSecret))
	if err != nil {
		return fmt.Errorf("create cert client: %w", err)
	}
	if err := retry.Do(ctx, func() error {
		return certClient.Certs.Import(ctx, getLabel, cert.CertificatePEM)
	}); err != nil {
		return fmt.Errorf("upload signer cert: %w", err)
	}
	// Small pause — IBM Verify needs a moment to index the new signer cert.
	time.Sleep(2 * time.Second)

	jwtID := fmt.Sprintf("%s-%d", getLabel, time.Now().UnixNano())
	jwt, err := crypto.GenerateSignedJWT(crypto.JWTRequest{
		Issuer:        getIssuer,
		Subject:       getSubject,
		KeyID:         getLabel,
		JWTID:         jwtID,
		PrivateKeyPEM: cert.PrivateKeyPEM,
		ExpiresIn:     getJWTExpiresIn,
	})
	if err != nil {
		return fmt.Errorf("sign JWT: %w", err)
	}

	stsClient, err := client.New(getTenant, client.WithClientCredentials(getSTSClientID, getSTSClientSecret))
	if err != nil {
		return fmt.Errorf("create STS client: %w", err)
	}
	var exchanged *client.ExchangeResult
	if err := retry.Do(ctx, func() error {
		var e error
		exchanged, e = stsClient.Token.Exchange(ctx, jwt.Token, getSubjectTokenType)
		return e
	}); err != nil {
		return fmt.Errorf("exchange token: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), exchanged.AccessToken)
	return nil
}
