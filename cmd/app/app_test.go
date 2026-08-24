package app

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

// fakeAppServer stubs the IBM Verify token and applications endpoints.
func fakeAppServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token",
			"expires_in":   3600,
		})
	})

	mux.HandleFunc("/v1.0/applications", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"_links": map[string]any{
					"self": map[string]any{"href": "/v1.0/applications/app-abc"},
				},
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"_embedded": map[string]any{
				"applications": []any{
					map[string]any{
						"id":               "app-abc",
						"name":             "My App",
						"applicationState": true,
					},
					map[string]any{
						"id":               "app-xyz",
						"name":             "Other App",
						"applicationState": false,
					},
				},
			},
		})
	})

	mux.HandleFunc("/v1.0/applications/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":               "app-abc",
				"name":             "My App",
				"applicationState": true,
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

func TestAppList_text(t *testing.T) {
	srv := fakeAppServer(t)
	defer srv.Close()

	listTenant, listClientID, listClientSecret = srv.URL, "cid", "csec"
	var buf bytes.Buffer
	listCmd.SetOut(&buf)

	if err := runListApps(listCmd, nil); err != nil {
		t.Fatalf("runListApps: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "app-abc") {
		t.Errorf("expected app-abc in output, got: %s", out)
	}
	if !strings.Contains(out, "My App") {
		t.Errorf("expected 'My App' in output, got: %s", out)
	}
}

func TestAppList_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v1.0/applications", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"_embedded": map[string]any{"applications": []any{}},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	listTenant, listClientID, listClientSecret = srv.URL, "cid", "csec"
	var buf bytes.Buffer
	listCmd.SetOut(&buf)

	if err := runListApps(listCmd, nil); err != nil {
		t.Fatalf("runListApps empty: %v", err)
	}
	if !strings.Contains(buf.String(), "No applications") {
		t.Errorf("expected 'No applications' message, got: %s", buf.String())
	}
}

func TestAppList_json(t *testing.T) {
	srv := fakeAppServer(t)
	defer srv.Close()

	listTenant, listClientID, listClientSecret = srv.URL, "cid", "csec"
	cmd.GlobalOutput = output.JSON
	defer func() { cmd.GlobalOutput = output.Text }()

	var buf bytes.Buffer
	listCmd.SetOut(&buf)

	if err := runListApps(listCmd, nil); err != nil {
		t.Fatalf("runListApps json: %v", err)
	}
	var list []interface{}
	if err := json.Unmarshal(buf.Bytes(), &list); err != nil {
		t.Fatalf("output is not valid JSON array: %v\noutput: %s", err, buf.String())
	}
	if len(list) != 2 {
		t.Errorf("expected 2 apps in JSON, got %d", len(list))
	}
}

func TestAppGet_text(t *testing.T) {
	srv := fakeAppServer(t)
	defer srv.Close()

	getTenant, getClientID, getClientSecret, getAppID = srv.URL, "cid", "csec", "app-abc"
	var buf bytes.Buffer
	getCmd.SetOut(&buf)

	if err := runGetApp(getCmd, nil); err != nil {
		t.Fatalf("runGetApp: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "app-abc") {
		t.Errorf("expected app-abc in output, got: %s", out)
	}
	if !strings.Contains(out, "My App") {
		t.Errorf("expected 'My App' in output, got: %s", out)
	}
}

func TestAppGet_json(t *testing.T) {
	srv := fakeAppServer(t)
	defer srv.Close()

	getTenant, getClientID, getClientSecret, getAppID = srv.URL, "cid", "csec", "app-abc"
	cmd.GlobalOutput = output.JSON
	defer func() { cmd.GlobalOutput = output.Text }()

	var buf bytes.Buffer
	getCmd.SetOut(&buf)

	if err := runGetApp(getCmd, nil); err != nil {
		t.Fatalf("runGetApp json: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if m["id"] != "app-abc" {
		t.Errorf("expected id=app-abc in JSON, got: %v", m["id"])
	}
}

func TestAppCreate_text(t *testing.T) {
	srv := fakeAppServer(t)
	defer srv.Close()

	createTenant, createClientID, createClientSecret = srv.URL, "cid", "csec"
	createName, createTemplateID = "My App", "tmpl-001"
	var buf bytes.Buffer
	createCmd.SetOut(&buf)

	if err := runCreateApp(createCmd, nil); err != nil {
		t.Fatalf("runCreateApp: %v", err)
	}
	if !strings.Contains(buf.String(), "created") {
		t.Errorf("expected 'created' in output, got: %s", buf.String())
	}
}

func TestAppDelete_ok(t *testing.T) {
	srv := fakeAppServer(t)
	defer srv.Close()

	deleteTenant, deleteClientID, deleteClientSecret, deleteAppID = srv.URL, "cid", "csec", "app-abc"
	var buf bytes.Buffer
	deleteCmd.SetOut(&buf)

	if err := runDeleteApp(deleteCmd, nil); err != nil {
		t.Fatalf("runDeleteApp: %v", err)
	}
	if !strings.Contains(buf.String(), "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", buf.String())
	}
}

func TestAppDelete_notFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v1.0/applications/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"messageID":          "CSIAK0001E",
			"messageDescription": "Not found",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	deleteTenant, deleteClientID, deleteClientSecret, deleteAppID = srv.URL, "cid", "csec", "no-such-app"

	if err := runDeleteApp(deleteCmd, nil); err == nil {
		t.Fatal("expected error for not-found app, got nil")
	}
}

func TestFmtField(t *testing.T) {
	m := map[string]interface{}{
		"strField":  "hello",
		"boolTrue":  true,
		"boolFalse": false,
	}
	if got := fmtField(m, "strField"); got != "hello" {
		t.Errorf("strField: got %q", got)
	}
	if got := fmtField(m, "boolTrue"); got != "true" {
		t.Errorf("boolTrue: got %q", got)
	}
	if got := fmtField(m, "boolFalse"); got != "false" {
		t.Errorf("boolFalse: got %q", got)
	}
	if got := fmtField(m, "missing"); got != "" {
		t.Errorf("missing field: got %q", got)
	}
}
