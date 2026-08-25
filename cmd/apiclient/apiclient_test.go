package apiclient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/internal/output"
)

// fakeAPIClientServer stubs the IBM Verify token and API clients endpoints.
// The SDK v1.6.4 apiclients package uses raw HTTP (not Fern-generated client)
// so the responses here match what the real IBM Verify API returns.
func fakeAPIClientServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// OAuth token endpoint
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token",
			"expires_in":   3600,
		})
	})

	// list + create — plain JSON array (no wrapper object)
	mux.HandleFunc("/v1.0/apiclients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"clientId":     "client-abc",
				"clientName":   "My Client",
				"clientSecret": "s3cr3t",
				"enabled":      true,
				"entitlements": []any{"manageAPIClients"},
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]any{
			map[string]any{
				"clientId":   "client-abc",
				"clientName": "My Client",
				"enabled":    true,
			},
			map[string]any{
				"clientId":   "client-xyz",
				"clientName": "Other Client",
				"enabled":    false,
			},
		})
	})

	// get + delete by id
	mux.HandleFunc("/v1.0/apiclients/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"clientId":   "client-abc",
				"clientName": "My Client",
				"enabled":    true,
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

func TestAPIClientList_noError(t *testing.T) {
	srv := fakeAPIClientServer(t)
	defer srv.Close()

	listTenant, listClientID, listClientSecret = srv.URL, "cid", "csec"
	var buf bytes.Buffer
	listCmd.SetOut(&buf)

	if err := runListAPIClients(listCmd, nil); err != nil {
		t.Fatalf("runListAPIClients returned unexpected error: %v", err)
	}
	// SDK v1.6.4 uses raw HTTP for List — results come back as parsed maps
	out := buf.String()
	if strings.Contains(out, "Error") || strings.Contains(out, "error") {
		t.Errorf("unexpected error text in output: %s", out)
	}
}

func TestAPIClientList_json(t *testing.T) {
	srv := fakeAPIClientServer(t)
	defer srv.Close()

	listTenant, listClientID, listClientSecret = srv.URL, "cid", "csec"
	cmd.GlobalOutput = output.JSON
	defer func() { cmd.GlobalOutput = output.Text }()

	var buf bytes.Buffer
	listCmd.SetOut(&buf)

	if err := runListAPIClients(listCmd, nil); err != nil {
		t.Fatalf("runListAPIClients json: %v", err)
	}
	// Must be a valid JSON array (may be empty or populated depending on SDK parse)
	var list []interface{}
	if err := json.Unmarshal(buf.Bytes(), &list); err != nil {
		t.Fatalf("output is not valid JSON array: %v\noutput: %s", err, buf.String())
	}
}

func TestAPIClientGet_text(t *testing.T) {
	srv := fakeAPIClientServer(t)
	defer srv.Close()

	getTenant, getClientID, getClientSecret, getAPIClientID = srv.URL, "cid", "csec", "client-abc"
	var buf bytes.Buffer
	getCmd.SetOut(&buf)

	if err := runGetAPIClient(getCmd, nil); err != nil {
		t.Fatalf("runGetAPIClient: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "client-abc") {
		t.Errorf("expected client-abc in output, got: %s", out)
	}
	if !strings.Contains(out, "My Client") {
		t.Errorf("expected 'My Client' in output, got: %s", out)
	}
}

func TestAPIClientGet_json(t *testing.T) {
	srv := fakeAPIClientServer(t)
	defer srv.Close()

	getTenant, getClientID, getClientSecret, getAPIClientID = srv.URL, "cid", "csec", "client-abc"
	cmd.GlobalOutput = output.JSON
	defer func() { cmd.GlobalOutput = output.Text }()

	var buf bytes.Buffer
	getCmd.SetOut(&buf)

	if err := runGetAPIClient(getCmd, nil); err != nil {
		t.Fatalf("runGetAPIClient json: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if m["clientId"] != "client-abc" {
		t.Errorf("expected clientId=client-abc in JSON, got: %v", m["clientId"])
	}
}

func TestAPIClientGet_notFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v1.0/apiclients/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Not found"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	getTenant, getClientID, getClientSecret, getAPIClientID = srv.URL, "cid", "csec", "no-such-client"

	if err := runGetAPIClient(getCmd, nil); err == nil {
		t.Fatal("expected error for not-found client, got nil")
	}
}

func TestAPIClientCreate_text(t *testing.T) {
	srv := fakeAPIClientServer(t)
	defer srv.Close()

	createTenant, createClientID, createClientSecret = srv.URL, "cid", "csec"
	createName, createEntitlements, createEnabled = "My Client", []string{"manageAPIClients"}, true
	var buf bytes.Buffer
	createCmd.SetOut(&buf)

	if err := runCreateAPIClient(createCmd, nil); err != nil {
		t.Fatalf("runCreateAPIClient: %v", err)
	}
	if !strings.Contains(buf.String(), "created") {
		t.Errorf("expected 'created' in output, got: %s", buf.String())
	}
}

func TestAPIClientCreate_json(t *testing.T) {
	srv := fakeAPIClientServer(t)
	defer srv.Close()

	createTenant, createClientID, createClientSecret = srv.URL, "cid", "csec"
	createName, createEntitlements, createEnabled = "My Client", []string{}, true
	cmd.GlobalOutput = output.JSON
	defer func() { cmd.GlobalOutput = output.Text }()

	var buf bytes.Buffer
	createCmd.SetOut(&buf)

	if err := runCreateAPIClient(createCmd, nil); err != nil {
		t.Fatalf("runCreateAPIClient json: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	// clientSecret must be present in the create response (one-time only)
	if m["clientSecret"] == nil {
		t.Errorf("expected clientSecret in create response, got nil")
	}
}

func TestAPIClientDelete_ok(t *testing.T) {
	srv := fakeAPIClientServer(t)
	defer srv.Close()

	deleteTenant, deleteClientID, deleteClientSecret, deleteAPIClientID = srv.URL, "cid", "csec", "client-abc"
	var buf bytes.Buffer
	deleteCmd.SetOut(&buf)

	if err := runDeleteAPIClient(deleteCmd, nil); err != nil {
		t.Fatalf("runDeleteAPIClient: %v", err)
	}
	if !strings.Contains(buf.String(), "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", buf.String())
	}
}

func TestAPIClientDelete_notFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v1.0/apiclients/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Not found"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	deleteTenant, deleteClientID, deleteClientSecret, deleteAPIClientID = srv.URL, "cid", "csec", "no-such-client"

	if err := runDeleteAPIClient(deleteCmd, nil); err == nil {
		t.Fatal("expected error for not-found client, got nil")
	}
}
