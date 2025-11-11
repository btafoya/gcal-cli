package auth

import (
	"encoding/json"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"github.com/btafoya/gcal-cli/pkg/types"
)

// TokenStorage handles secure storage and retrieval of OAuth2 tokens
type TokenStorage struct {
	TokenPath string
}

// NewTokenStorage creates a new token storage instance
func NewTokenStorage(tokenPath string) *TokenStorage {
	return &TokenStorage{
		TokenPath: tokenPath,
	}
}

// SaveToken saves a token to disk with secure permissions
func (ts *TokenStorage) SaveToken(token *oauth2.Token) error {
	// Ensure parent directory exists
	tokenDir := filepath.Dir(ts.TokenPath)
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		return types.ErrFileError.
			WithDetails("could not create token directory: " + tokenDir).
			WithWrappedError(err)
	}

	// Marshal token to JSON
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return types.ErrFileError.
			WithDetails("could not marshal token to JSON").
			WithWrappedError(err)
	}

	// Write token to file with restrictive permissions (0600 = rw-------)
	if err := os.WriteFile(ts.TokenPath, data, 0600); err != nil {
		return types.ErrFileError.
			WithDetails("could not write token file: " + ts.TokenPath).
			WithWrappedError(err)
	}

	return nil
}

// LoadToken loads a token from disk
func (ts *TokenStorage) LoadToken() (*oauth2.Token, error) {
	// Check if token file exists
	if _, err := os.Stat(ts.TokenPath); os.IsNotExist(err) {
		return nil, types.ErrAuthFailed("no authentication token found").
			WithSuggestedAction("Run 'gcal-cli auth login' to authenticate")
	}

	// Read token file
	data, err := os.ReadFile(ts.TokenPath)
	if err != nil {
		return nil, types.ErrFileError.
			WithDetails("could not read token file: " + ts.TokenPath).
			WithWrappedError(err)
	}

	// Unmarshal token
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, types.ErrFileError.
			WithDetails("could not parse token file").
			WithWrappedError(err).
			WithSuggestedAction("Token file may be corrupted. Try re-authenticating with 'gcal-cli auth login'")
	}

	return &token, nil
}

// DeleteToken removes the stored token
func (ts *TokenStorage) DeleteToken() error {
	// Check if file exists
	if _, err := os.Stat(ts.TokenPath); os.IsNotExist(err) {
		// Token doesn't exist, consider it already deleted
		return nil
	}

	// Remove token file
	if err := os.Remove(ts.TokenPath); err != nil {
		return types.ErrFileError.
			WithDetails("could not delete token file: " + ts.TokenPath).
			WithWrappedError(err)
	}

	return nil
}

// TokenExists checks if a token file exists
func (ts *TokenStorage) TokenExists() bool {
	_, err := os.Stat(ts.TokenPath)
	return err == nil
}

// ValidateTokenPermissions checks if the token file has secure permissions
func (ts *TokenStorage) ValidateTokenPermissions() error {
	fileInfo, err := os.Stat(ts.TokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, no permissions to validate
		}
		return types.ErrFileError.
			WithDetails("could not check token file permissions").
			WithWrappedError(err)
	}

	// Check if permissions are too open (should be 0600 or more restrictive)
	perm := fileInfo.Mode().Perm()
	if perm&0077 != 0 {
		// File is readable/writable by group or others
		return types.ErrConfigError("token file has insecure permissions").
			WithDetails(ts.TokenPath).
			WithSuggestedAction("Run 'chmod 600 " + ts.TokenPath + "' to fix permissions")
	}

	return nil
}
