package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/ChrisVerde02/ibmverify-go/crypto"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the full token-exchange flow (cert → upload → JWT → token)",
	Long: `Performs the complete IBM Verify token-exchange flow in one command:

  1. Generates a self-signed RSA certificate
  2. Obtains a cert-manager access token (client credentials)
  3. Uploads the certificate to IBM Verify as a signer cert
  4. Signs a short-lived JWT with the certificate's private key
  5. Exchanges the JWT for an IBM Verify access token
  6. Introspects the access token and prints the result

This is the CLI equivalent of running terraform apply against examples2.`,
	RunE: runFlow,
}

var (
	runTenant                  string
	runSTSClientID             string
	runSTSClientSecret         string
	runCertManagerClientID     string
	runCertManagerClientSecret string
	runSubject                 string
	runIssuer                  string
	runLabel                   string
	runOrganization            string
	runCountry                 string
	runValidityDays            int
	runKeySize                 int
	runSubjectTokenType        string
)

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVar(&runTenant, "tenant", "", "IBM Verify tenant URL (required)")
	runCmd.Flags().StringVar(&runSTSClientID, "sts-client-id", "", "STS client ID (required)")
	runCmd.Flags().StringVar(&runSTSClientSecret, "sts-client-secret", "", "STS client secret (required)")
	runCmd.Flags().StringVar(&runCertManagerClientID, "cert-manager-client-id", "", "Cert-manager client ID (required)")
	runCmd.Flags().StringVar(&runCertManagerClientSecret, "cert-manager-client-secret", "", "Cert-manager client secret (required)")
	runCmd.Flags().StringVar(&runSubject, "subject", "", "JWT subject — IBM Verify username (required)")
	runCmd.Flags().StringVar(&runIssuer, "issuer", "", "JWT issuer claim (required)")
	runCmd.Flags().StringVar(&runLabel, "label", "", "Signer cert label / JWT kid (required)")
	runCmd.Flags().StringVar(&runOrganization, "organization", "IBM", "Certificate O field")
	runCmd.Flags().StringVar(&runCountry, "country", "US", "Certificate C field (2-letter)")
	runCmd.Flags().IntVar(&runValidityDays, "validity-days", 365, "Certificate validity in days")
	runCmd.Flags().IntVar(&runKeySize, "key-size", 4096, "RSA key size (2048, 3072, or 4096)")
	runCmd.Flags().StringVar(&runSubjectTokenType, "subject-token-type", "urn:demo:token-type:user-jwt", "Subject token type URN")

	_ = runCmd.MarkFlagRequired("tenant")
	_ = runCmd.MarkFlagRequired("sts-client-id")
	_ = runCmd.MarkFlagRequired("sts-client-secret")
	_ = runCmd.MarkFlagRequired("cert-manager-client-id")
	_ = runCmd.MarkFlagRequired("cert-manager-client-secret")
	_ = runCmd.MarkFlagRequired("subject")
	_ = runCmd.MarkFlagRequired("issuer")
	_ = runCmd.MarkFlagRequired("label")
}

func runFlow(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Step 1 — generate self-signed certificate
	fmt.Print("  Generating certificate... ")
	cert, err := crypto.GenerateSelfSignedCertificate(crypto.CertificateRequest{
		CommonName:   runLabel,
		Organization: runOrganization,
		Country:      runCountry,
		ValidityDays: runValidityDays,
		KeySize:      runKeySize,
	})
	if err != nil {
		return fmt.Errorf("generate certificate: %w", err)
	}
	fmt.Printf("✓  (CN=%s, valid %d days)\n", runLabel, runValidityDays)

	// Step 2 — obtain cert-manager client credentials token
	fmt.Print("  Obtaining cert-manager token... ")
	certTokenResult, err := client.GetClientCredentialsToken(ctx, client.ClientCredentialsRequest{
		TenantURL:    runTenant,
		ClientID:     runCertManagerClientID,
		ClientSecret: runCertManagerClientSecret,
	})
	if err != nil {
		return fmt.Errorf("get cert-manager token: %w", err)
	}
	fmt.Println("✓")

	// Step 3 — upload certificate to IBM Verify
	fmt.Print("  Uploading signer certificate... ")
	if err := client.ImportSignerCert(ctx, client.SignerCertRequest{
		TenantURL:      runTenant,
		AccessToken:    certTokenResult.AccessToken,
		CertificatePEM: cert.CertificatePEM,
		Label:          runLabel,
	}); err != nil {
		return fmt.Errorf("upload signer cert: %w", err)
	}
	fmt.Printf("✓  (label=%s)\n", runLabel)

	// Step 4 — sign a short-lived JWT
	fmt.Print("  Signing JWT... ")
	jwtID := fmt.Sprintf("%s-%d", runLabel, time.Now().UnixNano())
	jwt, err := crypto.GenerateSignedJWT(crypto.JWTRequest{
		Issuer:        runIssuer,
		Subject:       runSubject,
		KeyID:         runLabel,
		JWTID:         jwtID,
		PrivateKeyPEM: cert.PrivateKeyPEM,
		ExpiresIn:     15 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("sign JWT: %w", err)
	}
	fmt.Printf("✓  (kid=%s, exp=15min)\n", runLabel)

	// Step 5 — exchange JWT for IBM Verify access token
	fmt.Print("  Exchanging token... ")
	exchanged, err := client.ExchangeToken(ctx, client.TokenExchangeRequest{
		TenantURL:        runTenant,
		ClientID:         runSTSClientID,
		ClientSecret:     runSTSClientSecret,
		SubjectToken:     jwt.Token,
		SubjectTokenType: runSubjectTokenType,
	})
	if err != nil {
		return fmt.Errorf("exchange token: %w", err)
	}
	fmt.Printf("✓  (expires in %ds)\n", exchanged.ExpiresIn)

	// Step 6 — introspect the access token
	fmt.Print("  Introspecting token... ")
	info, err := client.IntrospectToken(ctx, client.IntrospectionRequest{
		TenantURL:    runTenant,
		ClientID:     runSTSClientID,
		ClientSecret: runSTSClientSecret,
		Token:        exchanged.AccessToken,
	})
	if err != nil {
		return fmt.Errorf("introspect token: %w", err)
	}
	fmt.Printf("✓  (subject=%s, user=%s)\n\n", info.Subject, info.Username)

	fmt.Println("Access token:")
	fmt.Println(exchanged.AccessToken)

	return nil
}
