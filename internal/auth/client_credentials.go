// Package auth provides helpers for obtaining IBM Verify access tokens.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// clientCredentialsResponse is the JSON body returned by IBM Verify.
type clientCredentialsResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// GetClientCredentialsToken fetches an access token from IBM Verify using the
// OAuth 2.0 client credentials grant.  The token endpoint is:
//
//	POST /v1.0/endpoint/default/token
func GetClientCredentialsToken(
	ctx context.Context,
	tenantURL, clientID, clientSecret string,
) (string, error) {
	endpoint := strings.TrimRight(tenantURL, "/") + "/v1.0/endpoint/default/token"

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("create client credentials request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send client credentials request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read client credentials response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"IBM Verify client credentials failed with HTTP %d: %s",
			resp.StatusCode, string(body),
		)
	}

	var result clientCredentialsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("decode client credentials response: %w", err)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("IBM Verify response did not contain an access_token")
	}

	return result.AccessToken, nil
}
