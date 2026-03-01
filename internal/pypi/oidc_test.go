package pypi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestOIDCToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("audience") != "pypi" {
			t.Errorf("request missing audience=pypi, got %q", r.URL.Query().Get("audience"))
		}
		if r.Header.Get("Authorization") != "Bearer test-bearer-token" {
			t.Errorf("Authorization = %q, want %q",
				r.Header.Get("Authorization"), "Bearer test-bearer-token")
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"value": "oidc-jwt-token"})
	}))
	defer server.Close()

	token, err := requestOIDCToken(server.URL, "test-bearer-token")
	if err != nil {
		t.Fatalf("requestOIDCToken: %v", err)
	}
	if token != "oidc-jwt-token" {
		t.Errorf("token = %q, want %q", token, "oidc-jwt-token")
	}
}

func TestRequestOIDCToken_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	_, err := requestOIDCToken(server.URL, "bad-token")
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}

func TestRequestOIDCToken_EmptyTokenValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"value": ""})
	}))
	defer server.Close()

	_, err := requestOIDCToken(server.URL, "token")
	if err == nil {
		t.Fatal("expected error for empty token value, got nil")
	}
}

func TestRequestOIDCToken_InvalidURL(t *testing.T) {
	_, err := requestOIDCToken("://not-a-valid-url", "token")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestRequestOIDCToken_AppendAudienceToExistingQuery(t *testing.T) {
	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"value": "tok"})
	}))
	defer server.Close()

	_, err := requestOIDCToken(server.URL+"?existing=1", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedQuery, "audience=pypi") {
		t.Errorf("audience=pypi not in query %q", capturedQuery)
	}
	if !strings.Contains(capturedQuery, "existing=1") {
		t.Errorf("existing=1 not preserved in query %q", capturedQuery)
	}
}

func TestExchangeForUploadToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if body.Token != "oidc-jwt" {
			t.Errorf("request token = %q, want %q", body.Token, "oidc-jwt")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "pypi-upload-token"})
	}))
	defer server.Close()

	orig := pypiMintTokenURL
	pypiMintTokenURL = server.URL
	defer func() { pypiMintTokenURL = orig }()

	token, err := exchangeForUploadToken("oidc-jwt")
	if err != nil {
		t.Fatalf("exchangeForUploadToken: %v", err)
	}
	if token != "pypi-upload-token" {
		t.Errorf("token = %q, want %q", token, "pypi-upload-token")
	}
}

func TestExchangeForUploadToken_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	orig := pypiMintTokenURL
	pypiMintTokenURL = server.URL
	defer func() { pypiMintTokenURL = orig }()

	_, err := exchangeForUploadToken("oidc-jwt")
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}

func TestExchangeForUploadToken_EmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"token": ""})
	}))
	defer server.Close()

	orig := pypiMintTokenURL
	pypiMintTokenURL = server.URL
	defer func() { pypiMintTokenURL = orig }()

	_, err := exchangeForUploadToken("oidc-jwt")
	if err == nil {
		t.Fatal("expected error for empty upload token, got nil")
	}
}
