package token

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func resetGetFlags(cmd interface{ ResetFlags() }) {
	getCmd.ResetFlags()
	getCmd.Flags().StringVar(&getTenant, "tenant", "", "")
	getCmd.Flags().StringVar(&getSTSClientID, "sts-client-id", "", "")
	getCmd.Flags().StringVar(&getSTSClientSecret, "sts-client-secret", "", "")
	getCmd.Flags().StringVar(&getCertManagerClientID, "cert-manager-client-id", "", "")
	getCmd.Flags().StringVar(&getCertManagerClientSecret, "cert-manager-client-secret", "", "")
	getCmd.Flags().StringVar(&getSubject, "subject", "", "")
	getCmd.Flags().StringVar(&getIssuer, "issuer", "", "")
	getCmd.Flags().StringVar(&getLabel, "label", "", "")
	getCmd.Flags().StringVar(&getOrganization, "organization", "IBM", "")
	getCmd.Flags().StringVar(&getCountry, "country", "US", "")
	getCmd.Flags().IntVar(&getValidityDays, "validity-days", 365, "")
	getCmd.Flags().IntVar(&getKeySize, "key-size", 4096, "")
	getCmd.Flags().StringVar(&getSubjectTokenType, "subject-token-type", "urn:demo:token-type:user-jwt", "")
	getCmd.Flags().DurationVar(&getJWTExpiresIn, "jwt-expires-in", 15*time.Minute, "")
}

func baseFlags(srv *httptest.Server) []string {
	return []string{
		"--tenant", srv.URL,
		"--sts-client-id", "sts-id",
		"--sts-client-secret", "sts-secret",
		"--cert-manager-client-id", "cm-id",
		"--cert-manager-client-secret", "cm-secret",
		"--subject", "testuser",
		"--issuer", "https://test.ibm.com",
		"--label", "testlabel",
		"--key-size", "2048",
	}
}

func TestRunGet_outputsTokenOnly(t *testing.T) {
	srv := fakeVerify(t)
	defer srv.Close()

	resetGetFlags(getCmd)
	if err := getCmd.ParseFlags(baseFlags(srv)); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	var out bytes.Buffer
	getCmd.SetOut(&out)

	if err := runGet(getCmd, nil); err != nil {
		t.Fatalf("runGet returned error: %v", err)
	}

	got := strings.TrimSpace(out.String())
	if got != "exchanged-token-abc123" {
		t.Errorf("expected clean token on stdout, got %q", got)
	}
}

func TestRunGet_jwtExpiresInFlag(t *testing.T) {
	srv := fakeVerify(t)
	defer srv.Close()

	resetGetFlags(getCmd)
	flags := append(baseFlags(srv), "--jwt-expires-in", "5m")
	if err := getCmd.ParseFlags(flags); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	if getJWTExpiresIn != 5*time.Minute {
		t.Errorf("expected 5m, got %v", getJWTExpiresIn)
	}
	if err := runGet(getCmd, nil); err != nil {
		t.Fatalf("runGet returned error: %v", err)
	}
}

func TestRunGet_keySizeFlag(t *testing.T) {
	srv := fakeVerify(t)
	defer srv.Close()

	for _, size := range []string{"2048", "3072", "4096"} {
		t.Run("key-size="+size, func(t *testing.T) {
			resetGetFlags(getCmd)
			flags := append(baseFlags(srv), "--key-size", size)
			if err := getCmd.ParseFlags(flags); err != nil {
				t.Fatalf("parse flags: %v", err)
			}
			if err := runGet(getCmd, nil); err != nil {
				t.Fatalf("runGet returned error: %v", err)
			}
		})
	}
}

func TestRunGet_authError_returnsError(t *testing.T) {
	mux := http.NewServeMux()
	// client credentials returns 401
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_client"}`, http.StatusUnauthorized)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resetGetFlags(getCmd)
	if err := getCmd.ParseFlags(baseFlags(srv)); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	err := runGet(getCmd, nil)
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "cert client") && !strings.Contains(err.Error(), "upload") {
		t.Logf("error: %v", err)
	}
}

func TestRunGet_serverError_isRetried(t *testing.T) {
	attempts := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"access_token": "mgmt-token", "expires_in": 3600})
	})
	mux.HandleFunc("/v1.0/signercert", func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			http.Error(w, `{"error":"server_error"}`, http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"access_token": "exchanged-token-abc123", "expires_in": 3600})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resetGetFlags(getCmd)
	if err := getCmd.ParseFlags(baseFlags(srv)); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	if err := runGet(getCmd, nil); err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if attempts < 2 {
		t.Errorf("expected at least 2 upload attempts (retry), got %d", attempts)
	}
}
