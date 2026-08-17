package token

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunIntrospect_active(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"access_token": "mgmt-tok", "expires_in": 3600})
	})
	mux.HandleFunc("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"active":             true,
			"sub":                "user123",
			"preferred_username": "Bretton",
			"iss":                "https://tenant.verify.ibm.com/oauth2",
			"exp":                9999999999,
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	introspectTenant = srv.URL
	introspectClientID = "sts-id"
	introspectClientSecret = "sts-secret"
	introspectToken = "some-access-token"

	if err := runIntrospect(introspectCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunIntrospect_inactive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"access_token": "mgmt-tok", "expires_in": 3600})
	})
	mux.HandleFunc("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"active": false})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	introspectTenant = srv.URL
	introspectClientID = "sts-id"
	introspectClientSecret = "sts-secret"
	introspectToken = "expired-token"

	if err := runIntrospect(introspectCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunIntrospect_serverError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"access_token": "mgmt-tok", "expires_in": 3600})
	})
	mux.HandleFunc("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	introspectTenant = srv.URL
	introspectClientID = "sts-id"
	introspectClientSecret = "sts-secret"
	introspectToken = "bad-token"

	err := runIntrospect(introspectCmd, nil)
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
	if !strings.Contains(err.Error(), "introspect token") {
		t.Errorf("unexpected error: %v", err)
	}
}
