package user

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

// fakeUserServer stubs the IBM Verify token and SCIM users endpoints.
func fakeUserServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token",
			"expires_in":   3600,
		})
	})

	// list users (with optional filter query param)
	mux.HandleFunc("/v2.0/Users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/scim+json")
		filter := r.URL.Query().Get("filter")
		if filter != "" {
			// filtered — return only the matching user
			_ = json.NewEncoder(w).Encode(map[string]any{
				"totalResults": 1,
				"Resources": []any{
					map[string]any{
						"id":          "user-abc",
						"userName":    "john",
						"displayName": "John Doe",
						"active":      true,
					},
				},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"totalResults": 2,
			"Resources": []any{
				map[string]any{
					"id":          "user-abc",
					"userName":    "john",
					"displayName": "John Doe",
					"active":      true,
				},
				map[string]any{
					"id":          "user-xyz",
					"userName":    "jane",
					"displayName": "Jane Smith",
					"active":      false,
				},
			},
		})
	})

	// get single user
	mux.HandleFunc("/v2.0/Users/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/scim+json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "user-abc",
			"userName":    "john",
			"displayName": "John Doe",
			"active":      true,
		})
	})

	return httptest.NewServer(mux)
}

func TestUserList_text(t *testing.T) {
	srv := fakeUserServer(t)
	defer srv.Close()

	listTenant, listClientID, listClientSecret, listFilter = srv.URL, "cid", "csec", ""
	var buf bytes.Buffer
	listCmd.SetOut(&buf)

	if err := runListUsers(listCmd, nil); err != nil {
		t.Fatalf("runListUsers: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "user-abc") {
		t.Errorf("expected user-abc in output, got: %s", out)
	}
	if !strings.Contains(out, "john") {
		t.Errorf("expected 'john' in output, got: %s", out)
	}
	if !strings.Contains(out, "user-xyz") {
		t.Errorf("expected user-xyz in output, got: %s", out)
	}
}

func TestUserList_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v2.0/Users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/scim+json")
		_ = json.NewEncoder(w).Encode(map[string]any{"totalResults": 0, "Resources": []any{}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	listTenant, listClientID, listClientSecret, listFilter = srv.URL, "cid", "csec", ""
	var buf bytes.Buffer
	listCmd.SetOut(&buf)

	if err := runListUsers(listCmd, nil); err != nil {
		t.Fatalf("runListUsers empty: %v", err)
	}
	if !strings.Contains(buf.String(), "No users") {
		t.Errorf("expected 'No users' message, got: %s", buf.String())
	}
}

func TestUserList_filter(t *testing.T) {
	srv := fakeUserServer(t)
	defer srv.Close()

	listTenant, listClientID, listClientSecret = srv.URL, "cid", "csec"
	listFilter = `userName eq "john"`
	defer func() { listFilter = "" }()

	var buf bytes.Buffer
	listCmd.SetOut(&buf)

	if err := runListUsers(listCmd, nil); err != nil {
		t.Fatalf("runListUsers filter: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "user-abc") {
		t.Errorf("expected user-abc in filtered output, got: %s", out)
	}
	// jane should not appear
	if strings.Contains(out, "user-xyz") {
		t.Errorf("expected user-xyz absent from filtered output, got: %s", out)
	}
}

func TestUserList_json(t *testing.T) {
	srv := fakeUserServer(t)
	defer srv.Close()

	listTenant, listClientID, listClientSecret, listFilter = srv.URL, "cid", "csec", ""
	cmd.GlobalOutput = output.JSON
	defer func() { cmd.GlobalOutput = output.Text }()

	var buf bytes.Buffer
	listCmd.SetOut(&buf)

	if err := runListUsers(listCmd, nil); err != nil {
		t.Fatalf("runListUsers json: %v", err)
	}
	var list []interface{}
	if err := json.Unmarshal(buf.Bytes(), &list); err != nil {
		t.Fatalf("output is not valid JSON array: %v\noutput: %s", err, buf.String())
	}
	if len(list) != 2 {
		t.Errorf("expected 2 users in JSON, got %d", len(list))
	}
}

func TestUserGet_text(t *testing.T) {
	srv := fakeUserServer(t)
	defer srv.Close()

	getTenant, getClientID, getClientSecret, getUserID = srv.URL, "cid", "csec", "user-abc"
	var buf bytes.Buffer
	getCmd.SetOut(&buf)

	if err := runGetUser(getCmd, nil); err != nil {
		t.Fatalf("runGetUser: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "user-abc") {
		t.Errorf("expected user-abc in output, got: %s", out)
	}
	if !strings.Contains(out, "john") {
		t.Errorf("expected 'john' in output, got: %s", out)
	}
	if !strings.Contains(out, "John Doe") {
		t.Errorf("expected 'John Doe' in output, got: %s", out)
	}
}

func TestUserGet_json(t *testing.T) {
	srv := fakeUserServer(t)
	defer srv.Close()

	getTenant, getClientID, getClientSecret, getUserID = srv.URL, "cid", "csec", "user-abc"
	cmd.GlobalOutput = output.JSON
	defer func() { cmd.GlobalOutput = output.Text }()

	var buf bytes.Buffer
	getCmd.SetOut(&buf)

	if err := runGetUser(getCmd, nil); err != nil {
		t.Fatalf("runGetUser json: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if m["id"] != "user-abc" {
		t.Errorf("expected id=user-abc in JSON, got: %v", m["id"])
	}
}

func TestUserGet_notFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v2.0/Users/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "404",
			"detail": "User not found",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	getTenant, getClientID, getClientSecret, getUserID = srv.URL, "cid", "csec", "no-such-user"

	if err := runGetUser(getCmd, nil); err == nil {
		t.Fatal("expected error for not-found user, got nil")
	}
}

func TestUserCreate_text(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v2.0/Users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/scim+json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "user-new",
			"userName":    "newuser",
			"displayName": "New User",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	createTenant, createClientID, createClientSecret = srv.URL, "cid", "csec"
	createUserName, createPassword, createDisplayName = "newuser", "pass123", "New User"
	var buf bytes.Buffer
	createCmd.SetOut(&buf)

	if err := runCreateUser(createCmd, nil); err != nil {
		t.Fatalf("runCreateUser: %v", err)
	}
	if !strings.Contains(buf.String(), "created") {
		t.Errorf("expected 'created' in output, got: %s", buf.String())
	}
}

func TestUserDelete_ok(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v2.0/Users/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	deleteTenant, deleteClientID, deleteClientSecret, deleteUserID = srv.URL, "cid", "csec", "user-abc"
	var buf bytes.Buffer
	deleteCmd.SetOut(&buf)

	if err := runDeleteUser(deleteCmd, nil); err != nil {
		t.Fatalf("runDeleteUser: %v", err)
	}
	if !strings.Contains(buf.String(), "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", buf.String())
	}
}

func TestUserDelete_notFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/endpoint/default/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
	})
	mux.HandleFunc("/v2.0/Users/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "404", "detail": "not found"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	deleteTenant, deleteClientID, deleteClientSecret, deleteUserID = srv.URL, "cid", "csec", "no-such-user"

	if err := runDeleteUser(deleteCmd, nil); err == nil {
		t.Fatal("expected error for not-found user, got nil")
	}
}
