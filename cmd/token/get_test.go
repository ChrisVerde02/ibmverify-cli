package token

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeVerify stands up a minimal httptest server that handles the three
// endpoints the token-get flow hits: /token (client credentials),
// /v1.0/signercert (upload), and /oauth2/token (exchange).
func fakeVerify(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// client credentials
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "mgmt-token",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	})

	// upload signer cert
	mux.HandleFunc("/v1.0/signercert", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	// token exchange
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "exchanged-token-abc123",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	})

	return httptest.NewServer(mux)
}

func TestRunGet_outputsTokenOnly(t *testing.T) {
	srv := fakeVerify(t)
	defer srv.Close()

	tests := []struct {
		name  string
		flags []string
	}{
		{
			name: "all flags provided",
			flags: []string{
				"--tenant", srv.URL,
				"--sts-client-id", "sts-id",
				"--sts-client-secret", "sts-secret",
				"--cert-manager-client-id", "cm-id",
				"--cert-manager-client-secret", "cm-secret",
				"--subject", "testuser",
				"--issuer", "https://test.ibm.com",
				"--label", "testlabel",
				"--key-size", "2048",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := getCmd
			cmd.ResetFlags()
			// re-register flags
			cmd.Flags().StringVar(&getTenant, "tenant", "", "")
			cmd.Flags().StringVar(&getSTSClientID, "sts-client-id", "", "")
			cmd.Flags().StringVar(&getSTSClientSecret, "sts-client-secret", "", "")
			cmd.Flags().StringVar(&getCertManagerClientID, "cert-manager-client-id", "", "")
			cmd.Flags().StringVar(&getCertManagerClientSecret, "cert-manager-client-secret", "", "")
			cmd.Flags().StringVar(&getSubject, "subject", "", "")
			cmd.Flags().StringVar(&getIssuer, "issuer", "", "")
			cmd.Flags().StringVar(&getLabel, "label", "", "")
			cmd.Flags().StringVar(&getOrganization, "organization", "IBM", "")
			cmd.Flags().StringVar(&getCountry, "country", "US", "")
			cmd.Flags().IntVar(&getValidityDays, "validity-days", 365, "")
			cmd.Flags().IntVar(&getKeySize, "key-size", 4096, "")
			cmd.Flags().StringVar(&getSubjectTokenType, "subject-token-type", "urn:demo:token-type:user-jwt", "")

			if err := cmd.ParseFlags(tt.flags); err != nil {
				t.Fatalf("parse flags: %v", err)
			}
			if err := runGet(cmd, nil); err != nil {
				t.Fatalf("runGet returned error: %v", err)
			}
		})
	}
}
