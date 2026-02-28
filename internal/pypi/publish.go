package pypi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const pypiUploadURL = "https://upload.pypi.org/legacy/"

func Publish(cfg *Config) error {
	verb := "publishing"
	if cfg.DryRun {
		verb = "would publish"
	}

	fmt.Printf("pypi: %s version %s\n", verb, cfg.Version)

	token := ""
	if !cfg.DryRun {
		var err error
		token, err = mintToken()
		if err != nil {
			return err
		}
	}

	for _, a := range cfg.Artifacts {
		w, err := buildWheel(cfg, a)
		if err != nil {
			return fmt.Errorf("failed to build wheel for %s/%s: %w", a.Platform.GOOS, a.Platform.GOARCH, err)
		}
		fmt.Printf("pypi: %s %s...\n", verb, w.filename)
		if cfg.DryRun {
			continue
		}
		if err := uploadWheel(w, token); err != nil {
			return fmt.Errorf("pypi: failed to upload %s: %w", w.filename, err)
		}
	}

	fmt.Println("pypi: done")
	return nil
}

func uploadWheel(w wheelFile, token string) error {
	hash := sha256.Sum256(w.data)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	fields := [][2]string{
		{":action", "file_upload"},
		{"metadata_version", "2.1"},
		{"name", w.pkgName},
		{"version", w.version},
		{"requires_python", ">=3.7"},
		{"filetype", "bdist_wheel"},
		{"pyversion", "py3"},
		{"sha256_digest", hex.EncodeToString(hash[:])},
		{"protocol_version", "1"},
	}
	if w.summary != "" {
		fields = append(fields, [2]string{"summary", w.summary})
	}
	if w.license != "" {
		fields = append(fields, [2]string{"license", w.license})
	}
	if w.description != "" {
		fields = append(fields, [2]string{"description", w.description})
		if w.descContentType != "" {
			fields = append(fields, [2]string{"description_content_type", w.descContentType})
		}
	}
	for _, f := range fields {
		if err := mw.WriteField(f[0], f[1]); err != nil {
			return err
		}
	}

	fw, err := mw.CreateFormFile("content", w.filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, w.reader()); err != nil {
		return err
	}

	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, pypiUploadURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.SetBasicAuth("__token__", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(raw))
	}

	return nil
}
