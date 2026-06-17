package githubclient

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AppAuth handles GitHub App authentication using JWT
type AppAuth struct {
	AppID      int64
	PrivateKey *rsa.PrivateKey
	HTTPClient *http.Client
}

// NewAppAuth creates a new AppAuth instance
func NewAppAuth(appID int64, privateKeyPath string) (*AppAuth, error) {
	privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &AppAuth{
		AppID:      appID,
		PrivateKey: privateKey,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}, nil
}

// GenerateJWT creates a JWT token for GitHub App authentication
func (a *AppAuth) GenerateJWT() (string, error) {
	now := time.Now()
	expiresAt := now.Add(10 * time.Minute) // GitHub requires JWT to expire within 10 minutes

	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
		"iss": a.AppID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(a.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedToken, nil
}

// InstallationTokenResponse represents the response from GitHub's installation token endpoint
type InstallationTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetInstallationToken exchanges a JWT for an installation token
func (a *AppAuth) GetInstallationToken(installationID int64) (*InstallationTokenResponse, error) {
	jwtToken, err := a.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var tokenResp InstallationTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenResp, nil
}

// Installation represents a GitHub App installation
type Installation struct {
	ID        int64  `json:"id"`
	Account   Account `json:"account"`
	TargetType string `json:"target_type"`
}

// Account represents a GitHub account (user or organization)
type Account struct {
	Login string `json:"login"`
}

// ListInstallations lists all installations of the GitHub App
func (a *AppAuth) ListInstallations() ([]Installation, error) {
	jwtToken, err := a.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	url := "https://api.github.com/app/installations"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var installations []Installation
	if err := json.NewDecoder(resp.Body).Decode(&installations); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return installations, nil
}

// GetInstallationByOrg finds an installation for a specific organization
func (a *AppAuth) GetInstallationByOrg(orgName string) (*Installation, error) {
	installations, err := a.ListInstallations()
	if err != nil {
		return nil, err
	}

	for _, installation := range installations {
		if installation.Account.Login == orgName && installation.TargetType == "Organization" {
			return &installation, nil
		}
	}

	return nil, fmt.Errorf("no installation found for organization: %s", orgName)
}
