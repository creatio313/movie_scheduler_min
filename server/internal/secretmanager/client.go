package secretmanager

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DatabaseSecret represents the database secret stored in Secret Manager
type DatabaseSecret struct {
	DatabaseName string `json:"database_name"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

// SecretClient is a client for Sakura Cloud Secret Manager
type SecretClient struct {
	vaultID    string
	secretName string
	token      string
}

// NewSecretClient creates a new Secret Manager client
func NewSecretClient(vaultID, secretName string) (*SecretClient, error) {
	if vaultID == "" {
		return nil, fmt.Errorf("SAKURA_VAULT_ID is required")
	}

	if secretName == "" {
		secretName = "database_secret_value"
	}

	// Create Bearer token from access token and secret
	token, err := createBearerToken()
	if err != nil {
		return nil, fmt.Errorf("failed to create Bearer token: %w", err)
	}

	return &SecretClient{
		vaultID:    vaultID,
		secretName: secretName,
		token:      token,
	}, nil
}

// createBearerToken creates a Bearer token from SAKURA_ACCESS_TOKEN and SAKURA_ACCESS_TOKEN_SECRET
// Supports both environment variables and Docker secrets
func createBearerToken() (string, error) {
	var accessToken, accessTokenSecret string

	// Try to read from Docker secrets files first
	if accessTokenFile := os.Getenv("SAKURA_ACCESS_TOKEN_FILE"); accessTokenFile != "" {
		data, err := os.ReadFile(accessTokenFile)
		if err == nil {
			accessToken = string(bytes.TrimSpace(data))
		}
	}

	if accessTokenSecretFile := os.Getenv("SAKURA_ACCESS_TOKEN_SECRET_FILE"); accessTokenSecretFile != "" {
		data, err := os.ReadFile(accessTokenSecretFile)
		if err == nil {
			accessTokenSecret = string(bytes.TrimSpace(data))
		}
	}

	// Fallback to environment variables if secrets not found
	if accessToken == "" {
		accessToken = os.Getenv("SAKURA_ACCESS_TOKEN")
	}
	if accessTokenSecret == "" {
		accessTokenSecret = os.Getenv("SAKURA_ACCESS_TOKEN_SECRET")
	}

	if accessToken == "" || accessTokenSecret == "" {
		return "", fmt.Errorf("SAKURA_ACCESS_TOKEN and SAKURA_ACCESS_TOKEN_SECRET are required")
	}

	// Create Basic auth header: token:secret in base64
	credentials := accessToken + ":" + accessTokenSecret
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return "Basic " + encoded, nil
}

// FetchDatabaseSecret fetches the database secret from Secret Manager via API
func (sc *SecretClient) FetchDatabaseSecret() (*DatabaseSecret, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// API endpoint: https://secure.sakura.ad.jp/cloud/zone/is1c/api/cloud/1.1/secretmanager/vaults/{vaultID}/secrets/unveil
	apiURL := fmt.Sprintf("https://secure.sakura.ad.jp/cloud/zone/is1c/api/cloud/1.1/secretmanager/vaults/%s/secrets/unveil", sc.vaultID)

	// Prepare request body
	body := map[string]interface{}{
		"Secret": map[string]interface{}{
			"Name": sc.secretName,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", sc.token)
	req.Header.Set("Content-Type", "application/json")

	// Execute HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Secret Manager API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var apiResp struct {
		Secret struct {
			Name    string `json:"Name"`
			Version int    `json:"Version"`
			Value   string `json:"Value"`
		} `json:"Secret"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Parse the secret value as JSON
	var dbSecret DatabaseSecret
	if err := json.Unmarshal([]byte(apiResp.Secret.Value), &dbSecret); err != nil {
		return nil, fmt.Errorf("failed to parse secret value: invalid format")
	}

	return &dbSecret, nil
}
