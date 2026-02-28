package pypi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const pypiMintTokenURL = "https://pypi.org/oidc/mint-token/"

func mintToken() (string, error) {
	if token := os.Getenv("PYPI_TOKEN"); token != "" {
		fmt.Println("pypi: authenticating with PYPI_TOKEN")
		return token, nil
	}

	requestURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	requestToken := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")

	if requestURL == "" || requestToken == "" {
		return "", fmt.Errorf(
			"pypi: no credentials found\n" +
				"set PYPI_TOKEN for local publishing, or\n" +
				"ensure your workflow has 'id-token: write' permission and\n" +
				"a trusted publisher is registered at https://pypi.org/manage/account/publishing/",
		)
	}

	fmt.Println("pypi: authenticating with OIDC trusted publisher")
	oidcToken, err := requestOIDCToken(requestURL, requestToken)
	if err != nil {
		return "", fmt.Errorf("pypi: failed to request OIDC token: %w", err)
	}

	uploadToken, err := exchangeForUploadToken(oidcToken)
	if err != nil {
		return "", fmt.Errorf("pypi: failed to mint PyPI upload token: %w", err)
	}

	return uploadToken, nil
}

func requestOIDCToken(requestURL, requestToken string) (string, error) {
	u, err := url.Parse(requestURL)
	if err != nil {
		return "", fmt.Errorf("invalid OIDC request URL: %w", err)
	}
	q := u.Query()
	q.Set("audience", "pypi")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+requestToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Value == "" {
		return "", fmt.Errorf("OIDC token response was empty")
	}

	return result.Value, nil
}

func exchangeForUploadToken(oidcToken string) (string, error) {
	payload, err := json.Marshal(struct {
		Token string `json:"token"`
	}{Token: oidcToken})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(pypiMintTokenURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("PyPI mint-token returned status %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Token == "" {
		return "", fmt.Errorf("PyPI returned empty upload token")
	}

	return result.Token, nil
}
