package cmd

import (
	"context"
	gofmt "fmt"
	"os"
	"time"

	"github.com/ChrisVerde02/ibmverify-go/client"
	"github.com/ChrisVerde02/ibmverify-go/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ChrisVerde02/ibmverify-cli/internal/errkind"
	"github.com/ChrisVerde02/ibmverify-cli/internal/exitcode"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
	"github.com/ChrisVerde02/ibmverify-cli/internal/retry"
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

Progress is written to stderr. Stdout contains only the access token,
so TOKEN=$(ibmverify run ...) works cleanly.`,
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
	runJWTExpiresIn            time.Duration
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
	runCmd.Flags().DurationVar(&runJWTExpiresIn, "jwt-expires-in", 15*time.Minute, "JWT lifetime (e.g. 15m, 1h). Controls how long the signed JWT is valid before exchange.")

	_ = viper.BindPFlag("tenant", runCmd.Flags().Lookup("tenant"))
	_ = viper.BindPFlag("sts-client-id", runCmd.Flags().Lookup("sts-client-id"))
	_ = viper.BindPFlag("sts-client-secret", runCmd.Flags().Lookup("sts-client-secret"))
	_ = viper.BindPFlag("cert-manager-client-id", runCmd.Flags().Lookup("cert-manager-client-id"))
	_ = viper.BindPFlag("cert-manager-client-secret", runCmd.Flags().Lookup("cert-manager-client-secret"))
	_ = viper.BindPFlag("subject", runCmd.Flags().Lookup("subject"))
	_ = viper.BindPFlag("issuer", runCmd.Flags().Lookup("issuer"))
	_ = viper.BindPFlag("label", runCmd.Flags().Lookup("label"))
}

// progress writes a step message to stderr so stdout stays clean for data.
func progress(format string, a ...any) {
	gofmt.Fprintf(os.Stderr, format, a...)
}

func runFlow(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	tenant := viper.GetString("tenant")
	stsID := viper.GetString("sts-client-id")
	stsSecret := viper.GetString("sts-client-secret")
	certID := viper.GetString("cert-manager-client-id")
	certSecret := viper.GetString("cert-manager-client-secret")
	subject := viper.GetString("subject")
	issuer := viper.GetString("issuer")
	label := viper.GetString("label")

	if missing := requiredStrings(map[string]string{
		"--tenant / VERIFY_TENANT":                                  tenant,
		"--sts-client-id / VERIFY_STS_CLIENT_ID":                   stsID,
		"--sts-client-secret / VERIFY_STS_CLIENT_SECRET":           stsSecret,
		"--cert-manager-client-id / VERIFY_CERT_CLIENT_ID":         certID,
		"--cert-manager-client-secret / VERIFY_CERT_CLIENT_SECRET": certSecret,
		"--subject / VERIFY_SUBJECT":                                subject,
		"--issuer / VERIFY_ISSUER":                                  issuer,
		"--label / VERIFY_LABEL":                                    label,
	}); missing != "" {
		return gofmt.Errorf("missing required value(s): %s", missing)
	}

	fmt := outputFormat()

	// Step 1 — generate self-signed certificate
	progress("  Generating certificate... ")
	cert, err := crypto.GenerateSelfSignedCertificate(crypto.CertificateRequest{
		CommonName:   label,
		Organization: runOrganization,
		Country:      runCountry,
		ValidityDays: runValidityDays,
		KeySize:      runKeySize,
	})
	if err != nil {
		return cliError("generate certificate", err)
	}
	progress("✓  (CN=%s, valid %d days)\n", label, runValidityDays)

	// Step 2 — cert-manager client (auto-gets its own token internally)
	progress("  Obtaining cert-manager token... ")
	certClient, err := client.New(tenant, client.WithClientCredentials(certID, certSecret))
	if err != nil {
		return cliError("create cert client", err)
	}
	progress("✓\n")

	// Step 3 — upload certificate
	progress("  Uploading signer certificate... ")
	if err := retry.Do(ctx, func() error {
		return certClient.Certs.Import(ctx, label, cert.CertificatePEM)
	}); err != nil {
		return cliError("upload signer cert", err)
	}
	// Small pause — IBM Verify needs a moment to index the new signer cert
	// before it can validate JWTs signed with it.
	time.Sleep(2 * time.Second)
	progress("✓  (label=%s)\n", label)

	// Step 4 — sign a short-lived JWT
	progress("  Signing JWT... ")
	jwtID := gofmt.Sprintf("%s-%d", label, time.Now().UnixNano())
	jwt, err := crypto.GenerateSignedJWT(crypto.JWTRequest{
		Issuer:        issuer,
		Subject:       subject,
		KeyID:         label,
		JWTID:         jwtID,
		PrivateKeyPEM: cert.PrivateKeyPEM,
		ExpiresIn:     runJWTExpiresIn,
	})
	if err != nil {
		return cliError("sign JWT", err)
	}
	progress("✓  (kid=%s, exp=%s)\n", label, runJWTExpiresIn)

	// Step 5 — exchange JWT for access token
	progress("  Exchanging token... ")
	stsClient, err := client.New(tenant, client.WithClientCredentials(stsID, stsSecret))
	if err != nil {
		return cliError("create STS client", err)
	}
	var exchanged *client.ExchangeResult
	if err := retry.Do(ctx, func() error {
		var e error
		exchanged, e = stsClient.Token.Exchange(ctx, jwt.Token, runSubjectTokenType)
		return e
	}); err != nil {
		return cliError("exchange token", err)
	}
	progress("✓  (expires in %ds)\n", exchanged.ExpiresIn)

	// Step 6 — introspect the access token
	progress("  Introspecting token... ")
	info, err := stsClient.Token.Introspect(ctx, exchanged.AccessToken)
	if err != nil {
		return cliError("introspect token", err)
	}
	progress("✓  (subject=%s, user=%s)\n\n", info.Subject, info.Username)

	// Stdout = data only
	type runResult struct {
		AccessToken string `json:"access_token" yaml:"access_token"`
		ExpiresIn   int64  `json:"expires_in"   yaml:"expires_in"`
		Subject     string `json:"subject"      yaml:"subject"`
		Username    string `json:"username"     yaml:"username"`
	}
	result := runResult{
		AccessToken: exchanged.AccessToken,
		ExpiresIn:   exchanged.ExpiresIn,
		Subject:     info.Subject,
		Username:    info.Username,
	}
	if fmt == output.JSON || fmt == output.YAML {
		return output.Print(cmd.OutOrStdout(), fmt, result)
	}
	gofmt.Fprintln(cmd.OutOrStdout(), exchanged.AccessToken)
	return nil
}

func outputFormat() output.Format {
	f := output.Format(GlobalOutput)
	if !f.Valid() {
		return output.Text
	}
	return f
}

func cliError(step string, err error) error {
	if GlobalDebug {
		return gofmt.Errorf("%s: %w", step, err)
	}
	code := errkind.ExitCode(err)
	switch code {
	case exitcode.Auth:
		return gofmt.Errorf("%s: authentication failed (check client ID and secret)", step)
	case exitcode.NotFound:
		return gofmt.Errorf("%s: resource not found", step)
	case exitcode.RateLimit:
		return gofmt.Errorf("%s: rate limit exceeded — retry later", step)
	case exitcode.Server:
		return gofmt.Errorf("%s: IBM Verify server error — retry later", step)
	default:
		msg := err.Error()
		if i := indexAfterHTTPCode(msg); i >= 0 {
			msg = msg[:i]
		}
		return gofmt.Errorf("%s: %s", step, msg)
	}
}

func indexAfterHTTPCode(s string) int {
	for i := 0; i < len(s)-8; i++ {
		if s[i:i+5] == "HTTP " {
			for j := i + 5; j < len(s)-1; j++ {
				if s[j] == ':' && s[j+1] == ' ' {
					return j + 2
				}
			}
		}
	}
	return -1
}

func requiredStrings(m map[string]string) string {
	var missing []string
	for k, v := range m {
		if v == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) == 0 {
		return ""
	}
	out := ""
	for i, s := range missing {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}
