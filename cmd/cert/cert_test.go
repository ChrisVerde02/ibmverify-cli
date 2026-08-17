package cert

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeCertServer returns an httptest.Server that stubs the three IBM Verify
// cert-manager endpoints used by the cert subcommands.
func fakeCertServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// client credentials token
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "mgmt-token",
			"expires_in":   3600,
		})
	})

	// upload cert
	mux.HandleFunc("/v1.0/signercert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
	})

	// get / delete cert by label
	mux.HandleFunc("/v1.0/signercert/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"label":    "testlabel",
				"cert":     "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t",
				"subjectDN": "CN=testlabel",
				"issuerDN":  "CN=testlabel",
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	return httptest.NewServer(mux)
}

func TestCertGet_found(t *testing.T) {
	srv := fakeCertServer(t)
	defer srv.Close()

	getCertTenant = srv.URL
	getCertClientID = "cm-id"
	getCertClientSecret = "cm-secret"
	getCertLabel = "testlabel"

	var buf bytes.Buffer
	getCmd.SetOut(&buf)

	if err := runGetCert(getCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	// stdout is unused in runGetCert (uses fmt.Printf) — just check no error
	_ = out
}

func TestCertGet_notFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v1.0/signercert/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"messageId":"CSIAO5401E","messageDescription":"not found"}`, http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	getCertTenant = srv.URL
	getCertClientID = "cm-id"
	getCertClientSecret = "cm-secret"
	getCertLabel = "missing"

	if err := runGetCert(getCmd, nil); err != nil {
		t.Fatalf("expected nil error for not-found, got: %v", err)
	}
}

func TestCertDelete_success(t *testing.T) {
	srv := fakeCertServer(t)
	defer srv.Close()

	deleteTenant = srv.URL
	deleteClientID = "cm-id"
	deleteClientSecret = "cm-secret"
	deleteLabel = "testlabel"

	if err := runDelete(deleteCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCertDelete_notFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v1.0/signercert/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"messageId":"CSIAO5401E","messageDescription":"not found"}`, http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	deleteTenant = srv.URL
	deleteClientID = "cm-id"
	deleteClientSecret = "cm-secret"
	deleteLabel = "missing"

	err := runDelete(deleteCmd, nil)
	if err == nil {
		t.Fatal("expected error for not-found delete, got nil")
	}
	if !strings.Contains(err.Error(), "delete signer cert") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCertUpload_success(t *testing.T) {
	srv := fakeCertServer(t)
	defer srv.Close()

	// write a temporary PEM file
	tmp := filepath.Join(t.TempDir(), "cert.pem")
	if err := os.WriteFile(tmp, []byte("-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----\n"), 0600); err != nil {
		t.Fatal(err)
	}

	uploadTenant = srv.URL
	uploadClientID = "cm-id"
	uploadClientSecret = "cm-secret"
	uploadCertFile = tmp
	uploadLabel = "testlabel"

	if err := runUpload(uploadCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCertUpload_missingFile(t *testing.T) {
	srv := fakeCertServer(t)
	defer srv.Close()

	uploadTenant = srv.URL
	uploadClientID = "cm-id"
	uploadClientSecret = "cm-secret"
	uploadCertFile = "/nonexistent/cert.pem"
	uploadLabel = "testlabel"

	err := runUpload(uploadCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "read cert file") {
		t.Errorf("unexpected error: %v", err)
	}
}
