package pypi

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadWheel_Success(t *testing.T) {
	data := []byte("fake wheel zip content")
	hash := sha256.Sum256(data)
	expectedDigest := hex.EncodeToString(hash[:])

	var (
		receivedUser   string
		receivedPass   string
		receivedFields = make(map[string]string)
		receivedFile   []byte
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		receivedUser, receivedPass, ok = r.BasicAuth()
		if !ok {
			t.Error("request missing Basic Auth")
		}

		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
			t.Errorf("expected multipart Content-Type, got %q", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("reading multipart part: %v", err)
				break
			}
			b, _ := io.ReadAll(part)
			if part.FormName() == "content" {
				receivedFile = b
			} else {
				receivedFields[part.FormName()] = string(b)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	orig := pypiUploadURL
	pypiUploadURL = server.URL
	defer func() { pypiUploadURL = orig }()

	wf := wheelFile{
		filename: "mytool-1.0.0-py3-none-linux_x86_64.whl",
		pkgName:  "mytool",
		version:  "1.0.0",
		summary:  "A test tool",
		license:  "MIT",
		data:     data,
	}

	if err := uploadWheel(wf, "secret-token"); err != nil {
		t.Fatalf("uploadWheel: %v", err)
	}

	if receivedUser != "__token__" {
		t.Errorf("auth user = %q, want %q", receivedUser, "__token__")
	}
	if receivedPass != "secret-token" {
		t.Errorf("auth pass = %q, want %q", receivedPass, "secret-token")
	}
	if receivedFields["name"] != "mytool" {
		t.Errorf("name = %q, want %q", receivedFields["name"], "mytool")
	}
	if receivedFields["version"] != "1.0.0" {
		t.Errorf("version = %q, want %q", receivedFields["version"], "1.0.0")
	}
	if receivedFields["sha256_digest"] != expectedDigest {
		t.Errorf("sha256_digest = %q, want %q", receivedFields["sha256_digest"], expectedDigest)
	}
	if receivedFields[":action"] != "file_upload" {
		t.Errorf(":action = %q, want %q", receivedFields[":action"], "file_upload")
	}
	if receivedFields["summary"] != "A test tool" {
		t.Errorf("summary = %q, want %q", receivedFields["summary"], "A test tool")
	}
	if receivedFields["license"] != "MIT" {
		t.Errorf("license = %q, want %q", receivedFields["license"], "MIT")
	}
	if string(receivedFile) != string(data) {
		t.Errorf("file content mismatch: got %d bytes, want %d", len(receivedFile), len(data))
	}
}

func TestUploadWheel_201Created(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	orig := pypiUploadURL
	pypiUploadURL = server.URL
	defer func() { pypiUploadURL = orig }()

	wf := wheelFile{filename: "x.whl", pkgName: "x", version: "1.0.0", data: []byte("data")}
	if err := uploadWheel(wf, "tok"); err != nil {
		t.Fatalf("expected 201 to be treated as success, got: %v", err)
	}
}

func TestUploadWheel_OptionalFieldsOmitted(t *testing.T) {
	receivedFields := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			b, _ := io.ReadAll(part)
			receivedFields[part.FormName()] = string(b)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	orig := pypiUploadURL
	pypiUploadURL = server.URL
	defer func() { pypiUploadURL = orig }()

	wf := wheelFile{filename: "x.whl", pkgName: "x", version: "1.0.0", data: []byte("d")}
	if err := uploadWheel(wf, "tok"); err != nil {
		t.Fatalf("uploadWheel: %v", err)
	}

	if _, ok := receivedFields["summary"]; ok {
		t.Error("summary should not be sent when empty")
	}
	if _, ok := receivedFields["license"]; ok {
		t.Error("license should not be sent when empty")
	}
	if _, ok := receivedFields["description"]; ok {
		t.Error("description should not be sent when empty")
	}
}

func TestUploadWheel_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusConflict)
	}))
	defer server.Close()

	orig := pypiUploadURL
	pypiUploadURL = server.URL
	defer func() { pypiUploadURL = orig }()

	wf := wheelFile{filename: "x.whl", pkgName: "x", version: "1.0.0", data: []byte("d")}
	err := uploadWheel(wf, "tok")
	if err == nil {
		t.Fatal("expected error for non-200/201 status, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want 'already exists'", err.Error())
	}
}

func TestPypiError(t *testing.T) {
	tests := []struct {
		status  int
		wantSub string
	}{
		{http.StatusUnauthorized, "not authenticated"},
		{http.StatusForbidden, "permission denied"},
		{http.StatusConflict, "already exists"},
		{http.StatusRequestEntityTooLarge, "too large"},
		{http.StatusBadRequest, "invalid package"},
		{http.StatusInternalServerError, "unexpected status 500"},
		{http.StatusServiceUnavailable, "unexpected status 503"},
		{http.StatusGatewayTimeout, "unexpected status 504"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			got := pypiError(tt.status)
			if !strings.Contains(got, tt.wantSub) {
				t.Errorf("pypiError(%d) = %q, want substring %q", tt.status, got, tt.wantSub)
			}
		})
	}
}
