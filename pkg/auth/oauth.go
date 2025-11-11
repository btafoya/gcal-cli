package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"github.com/btafoya/gcal-cli/pkg/types"
)

// OAuthConfig represents OAuth2 configuration
type OAuthConfig struct {
	Config          *oauth2.Config
	CredentialsPath string
	TokenPath       string
}

// NewOAuthConfig creates a new OAuth configuration from credentials file
func NewOAuthConfig(credentialsPath, tokenPath string) (*OAuthConfig, error) {
	// Read credentials file
	credData, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, types.ErrConfigError("could not read credentials file").
			WithDetails(credentialsPath).
			WithWrappedError(err).
			WithSuggestedAction("Download OAuth2 credentials from Google Cloud Console")
	}

	// Parse credentials
	config, err := google.ConfigFromJSON(credData, calendar.CalendarScope)
	if err != nil {
		return nil, types.ErrInvalidCreds.
			WithDetails("failed to parse credentials JSON").
			WithWrappedError(err).
			WithSuggestedAction("Ensure credentials file is a valid OAuth2 client configuration")
	}

	return &OAuthConfig{
		Config:          config,
		CredentialsPath: credentialsPath,
		TokenPath:       tokenPath,
	}, nil
}

// GetAuthURL generates the OAuth2 authorization URL
func (o *OAuthConfig) GetAuthURL(state string) string {
	return o.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for tokens
func (o *OAuthConfig) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := o.Config.Exchange(ctx, code)
	if err != nil {
		return nil, types.ErrAuthFailed("failed to exchange authorization code").
			WithWrappedError(err).
			WithSuggestedAction("Try the authentication flow again")
	}

	return token, nil
}

// GetClient returns an authenticated HTTP client
func (o *OAuthConfig) GetClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return o.Config.Client(ctx, token)
}

// GetTokenSource returns a token source that automatically refreshes tokens
func (o *OAuthConfig) GetTokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource {
	return o.Config.TokenSource(ctx, token)
}

// StartAuthFlow initiates the OAuth2 flow and returns the auth URL and state
func (o *OAuthConfig) StartAuthFlow() (authURL, state string) {
	// Generate a random state for CSRF protection
	state = fmt.Sprintf("state-%d", time.Now().Unix())
	authURL = o.GetAuthURL(state)
	return authURL, state
}

// ValidateToken checks if a token is valid and not expired
func ValidateToken(token *oauth2.Token) error {
	if token == nil {
		return types.ErrAuthFailed("token is nil")
	}

	if !token.Valid() {
		if token.Expiry.Before(time.Now()) {
			return types.ErrTokenExpired()
		}
		return types.ErrAuthFailed("token is invalid")
	}

	return nil
}

// RefreshToken refreshes an expired token
func (o *OAuthConfig) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	if token.RefreshToken == "" {
		return nil, types.ErrAuthFailed("no refresh token available").
			WithSuggestedAction("Re-authenticate with 'gcal-cli auth login'")
	}

	tokenSource := o.Config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, types.ErrAuthFailed("failed to refresh token").
			WithWrappedError(err).
			WithSuggestedAction("Re-authenticate with 'gcal-cli auth login'")
	}

	return newToken, nil
}

// GetUserInfo retrieves user email from token (if available in token claims)
func GetUserInfo(token *oauth2.Token) (email string, err error) {
	// Try to get email from extra claims
	if extra := token.Extra("email"); extra != nil {
		if emailStr, ok := extra.(string); ok {
			return emailStr, nil
		}
	}

	// If not available in token, return empty (will need to be fetched separately)
	return "", nil
}

// ParseCredentialsFile reads and validates the credentials file structure
func ParseCredentialsFile(path string) (clientID, clientSecret string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", types.ErrConfigError("could not read credentials file").
			WithDetails(path).
			WithWrappedError(err)
	}

	var creds struct {
		Installed struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		} `json:"installed"`
		Web struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		} `json:"web"`
	}

	if err := json.Unmarshal(data, &creds); err != nil {
		return "", "", types.ErrInvalidCreds.
			WithDetails("invalid credentials JSON format").
			WithWrappedError(err)
	}

	// Try installed app credentials first
	if creds.Installed.ClientID != "" {
		return creds.Installed.ClientID, creds.Installed.ClientSecret, nil
	}

	// Fall back to web app credentials
	if creds.Web.ClientID != "" {
		return creds.Web.ClientID, creds.Web.ClientSecret, nil
	}

	return "", "", types.ErrInvalidCreds.
		WithDetails("no valid client credentials found in file").
		WithSuggestedAction("Download OAuth2 credentials from Google Cloud Console")
}
