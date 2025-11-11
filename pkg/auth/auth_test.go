package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// Test helpers

// createTestCredentials creates a temporary OAuth credentials file
func createTestCredentials(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")

	creds := map[string]interface{}{
		"installed": map[string]interface{}{
			"client_id":     "test-client-id",
			"client_secret": "test-client-secret",
			"redirect_uris": []string{"http://localhost:8080/oauth/callback"},
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
		},
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("Failed to marshal credentials: %v", err)
	}

	if err := os.WriteFile(credPath, data, 0600); err != nil {
		t.Fatalf("Failed to write credentials file: %v", err)
	}

	return credPath
}

// createTestToken creates a test OAuth2 token
func createTestToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
}

// createExpiredToken creates an expired OAuth2 token
func createExpiredToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}
}

// TestNewOAuthConfig tests OAuth configuration creation
func TestNewOAuthConfig(t *testing.T) {
	credPath := createTestCredentials(t)
	tokenPath := filepath.Join(t.TempDir(), "token.json")

	config, err := NewOAuthConfig(credPath, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuthConfig failed: %v", err)
	}

	if config.Config.ClientID != "test-client-id" {
		t.Errorf("Expected client ID 'test-client-id', got '%s'", config.Config.ClientID)
	}

	if config.Config.ClientSecret != "test-client-secret" {
		t.Errorf("Expected client secret 'test-client-secret', got '%s'", config.Config.ClientSecret)
	}

	if config.CredentialsPath != credPath {
		t.Errorf("Expected credentials path '%s', got '%s'", credPath, config.CredentialsPath)
	}
}

// TestNewOAuthConfig_InvalidFile tests error handling for invalid credentials
func TestNewOAuthConfig_InvalidFile(t *testing.T) {
	tokenPath := filepath.Join(t.TempDir(), "token.json")

	_, err := NewOAuthConfig("/nonexistent/credentials.json", tokenPath)
	if err == nil {
		t.Fatal("Expected error for nonexistent credentials file, got nil")
	}
}

// TestStartAuthFlow tests auth flow URL generation
func TestStartAuthFlow(t *testing.T) {
	credPath := createTestCredentials(t)
	tokenPath := filepath.Join(t.TempDir(), "token.json")

	config, err := NewOAuthConfig(credPath, tokenPath)
	if err != nil {
		t.Fatalf("NewOAuthConfig failed: %v", err)
	}

	authURL, state := config.StartAuthFlow()

	if authURL == "" {
		t.Error("Expected non-empty auth URL")
	}

	if state == "" {
		t.Error("Expected non-empty state")
	}

	// Verify state format
	if len(state) < 10 {
		t.Errorf("State too short, expected at least 10 chars, got %d", len(state))
	}
}

// TestValidateToken tests token validation
func TestValidateToken(t *testing.T) {
	tests := []struct {
		name    string
		token   *oauth2.Token
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   createTestToken(),
			wantErr: false,
		},
		{
			name:    "nil token",
			token:   nil,
			wantErr: true,
		},
		{
			name:    "expired token",
			token:   createExpiredToken(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTokenStorage tests token storage operations
func TestTokenStorage(t *testing.T) {
	tokenPath := filepath.Join(t.TempDir(), "token.json")
	storage := NewTokenStorage(tokenPath)
	testToken := createTestToken()

	// Test saving token
	if err := storage.SaveToken(testToken); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	// Verify token file exists
	if !storage.TokenExists() {
		t.Error("TokenExists() returned false after SaveToken()")
	}

	// Test loading token
	loadedToken, err := storage.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken failed: %v", err)
	}

	if loadedToken.AccessToken != testToken.AccessToken {
		t.Errorf("Expected access token '%s', got '%s'",
			testToken.AccessToken, loadedToken.AccessToken)
	}

	// Test token permissions
	if err := storage.ValidateTokenPermissions(); err != nil {
		t.Errorf("ValidateTokenPermissions failed: %v", err)
	}

	// Test deleting token
	if err := storage.DeleteToken(); err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}

	// Verify token file doesn't exist
	if storage.TokenExists() {
		t.Error("TokenExists() returned true after DeleteToken()")
	}
}

// TestTokenStorage_LoadNonexistent tests loading nonexistent token
func TestTokenStorage_LoadNonexistent(t *testing.T) {
	tokenPath := filepath.Join(t.TempDir(), "nonexistent.json")
	storage := NewTokenStorage(tokenPath)

	_, err := storage.LoadToken()
	if err == nil {
		t.Fatal("Expected error when loading nonexistent token, got nil")
	}
}

// TestCallbackServer tests callback server operations
func TestCallbackServer(t *testing.T) {
	server := NewCallbackServer(0, "test-state") // port 0 = random port

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	if err := server.Start(ctx); err != nil {
		t.Fatalf("Server.Start failed: %v", err)
	}
	defer server.Shutdown(ctx)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test successful callback
	go func() {
		req := httptest.NewRequest("GET",
			"/oauth/callback?code=test-code&state=test-state", nil)
		w := httptest.NewRecorder()
		server.handleCallback(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	}()

	// Wait for code
	code, err := server.WaitForCode(1 * time.Second)
	if err != nil {
		t.Fatalf("WaitForCode failed: %v", err)
	}

	if code != "test-code" {
		t.Errorf("Expected code 'test-code', got '%s'", code)
	}
}

// TestCallbackServer_InvalidState tests CSRF protection
func TestCallbackServer_InvalidState(t *testing.T) {
	server := NewCallbackServer(0, "valid-state")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Server.Start failed: %v", err)
	}
	defer server.Shutdown(ctx)

	// Test invalid state
	go func() {
		req := httptest.NewRequest("GET",
			"/oauth/callback?code=test-code&state=wrong-state", nil)
		w := httptest.NewRecorder()
		server.handleCallback(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	}()

	// Should receive error
	_, err := server.WaitForCode(1 * time.Second)
	if err == nil {
		t.Fatal("Expected error for invalid state, got nil")
	}
}

// TestCallbackServer_Error tests OAuth error handling
func TestCallbackServer_Error(t *testing.T) {
	server := NewCallbackServer(0, "test-state")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Server.Start failed: %v", err)
	}
	defer server.Shutdown(ctx)

	// Test OAuth error
	go func() {
		req := httptest.NewRequest("GET",
			"/oauth/callback?error=access_denied&error_description=User%20denied%20access", nil)
		w := httptest.NewRecorder()
		server.handleCallback(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	}()

	// Should receive error
	_, err := server.WaitForCode(1 * time.Second)
	if err == nil {
		t.Fatal("Expected error for OAuth error, got nil")
	}
}

// TestCallbackServer_Timeout tests timeout handling
func TestCallbackServer_Timeout(t *testing.T) {
	server := NewCallbackServer(0, "test-state")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Server.Start failed: %v", err)
	}
	defer server.Shutdown(ctx)

	// Don't send any callback, should timeout
	_, err := server.WaitForCode(100 * time.Millisecond)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

// TestGetUserInfo tests user info extraction
func TestGetUserInfo(t *testing.T) {
	// Test token without email
	token := createTestToken()
	email, err := GetUserInfo(token)
	if err != nil {
		t.Errorf("GetUserInfo failed: %v", err)
	}
	if email != "" {
		t.Errorf("Expected empty email for token without claims, got '%s'", email)
	}

	// Test token with email would require mocking oauth2 extra claims
	// which is complex, so we skip that test case
}

// TestParseCredentialsFile tests credentials file parsing
func TestParseCredentialsFile(t *testing.T) {
	credPath := createTestCredentials(t)

	clientID, clientSecret, err := ParseCredentialsFile(credPath)
	if err != nil {
		t.Fatalf("ParseCredentialsFile failed: %v", err)
	}

	if clientID != "test-client-id" {
		t.Errorf("Expected client ID 'test-client-id', got '%s'", clientID)
	}

	if clientSecret != "test-client-secret" {
		t.Errorf("Expected client secret 'test-client-secret', got '%s'", clientSecret)
	}
}

// TestParseCredentialsFile_Invalid tests error handling for invalid files
func TestParseCredentialsFile_Invalid(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "invalid.json")

	// Create invalid JSON file
	if err := os.WriteFile(credPath, []byte("not json"), 0600); err != nil {
		t.Fatalf("Failed to write invalid credentials: %v", err)
	}

	_, _, err := ParseCredentialsFile(credPath)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}
