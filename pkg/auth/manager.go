package auth

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/calendar/v3"
	"github.com/btafoya/gcal-cli/pkg/types"
)

// Manager handles authentication operations
type Manager struct {
	OAuth   *OAuthConfig
	Storage *TokenStorage
}

// NewManager creates a new authentication manager
func NewManager(credentialsPath, tokenPath string) (*Manager, error) {
	oauth, err := NewOAuthConfig(credentialsPath, tokenPath)
	if err != nil {
		return nil, err
	}

	storage := NewTokenStorage(tokenPath)

	return &Manager{
		OAuth:   oauth,
		Storage: storage,
	}, nil
}

// Login performs the OAuth2 login flow
func (m *Manager) Login(ctx context.Context) (*oauth2.Token, error) {
	// Generate auth URL and state
	authURL, state := m.OAuth.StartAuthFlow()

	// Start callback server
	server := NewCallbackServer(DefaultCallbackPort, state)
	if err := server.Start(ctx); err != nil {
		return nil, err
	}
	defer server.Shutdown(ctx)

	// Open browser to auth URL
	fmt.Printf("Opening browser for authentication...\n")
	fmt.Printf("If the browser doesn't open automatically, visit:\n%s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser automatically: %v\n", err)
		fmt.Printf("Please open the URL manually in your browser.\n\n")
	}

	// Wait for authorization code (timeout after 5 minutes)
	code, err := server.WaitForCode(5 * time.Minute)
	if err != nil {
		return nil, err
	}

	// Exchange code for token
	token, err := m.OAuth.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Save token
	if err := m.Storage.SaveToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

// Logout removes stored authentication
func (m *Manager) Logout() error {
	return m.Storage.DeleteToken()
}

// GetToken retrieves the stored token and refreshes it if necessary
func (m *Manager) GetToken(ctx context.Context) (*oauth2.Token, error) {
	// Load token
	token, err := m.Storage.LoadToken()
	if err != nil {
		return nil, err
	}

	// Check if token needs refresh
	if !token.Valid() {
		// Try to refresh
		refreshedToken, err := m.OAuth.RefreshToken(ctx, token)
		if err != nil {
			return nil, err
		}

		// Save refreshed token
		if err := m.Storage.SaveToken(refreshedToken); err != nil {
			return nil, err
		}

		token = refreshedToken
	}

	return token, nil
}

// GetCalendarService returns an authenticated Google Calendar service
func (m *Manager) GetCalendarService(ctx context.Context) (*calendar.Service, error) {
	token, err := m.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	client := m.OAuth.GetClient(ctx, token)

	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, types.ErrAPIError.
			WithDetails("failed to create calendar service").
			WithWrappedError(err)
	}

	return service, nil
}

// CheckAuthStatus checks the current authentication status
func (m *Manager) CheckAuthStatus(ctx context.Context) (authenticated bool, email string, expiresAt time.Time, err error) {
	// Check if token exists
	if !m.Storage.TokenExists() {
		return false, "", time.Time{}, nil
	}

	// Validate token file permissions
	if err := m.Storage.ValidateTokenPermissions(); err != nil {
		return false, "", time.Time{}, err
	}

	// Load token
	token, err := m.Storage.LoadToken()
	if err != nil {
		return false, "", time.Time{}, err
	}

	// Check token validity
	if err := ValidateToken(token); err != nil {
		// Token exists but is invalid/expired
		if appErr, ok := err.(*types.AppError); ok && appErr.Code == types.ErrCodeTokenExpired {
			// Try to refresh
			refreshedToken, refreshErr := m.OAuth.RefreshToken(ctx, token)
			if refreshErr != nil {
				return true, "", token.Expiry, refreshErr
			}

			// Save refreshed token
			if saveErr := m.Storage.SaveToken(refreshedToken); saveErr != nil {
				return true, "", token.Expiry, saveErr
			}

			token = refreshedToken
		} else {
			return true, "", token.Expiry, err
		}
	}

	// Get user email if available
	email, _ = GetUserInfo(token)

	// Try to get email from Calendar API if not in token
	if email == "" {
		service, err := m.GetCalendarService(ctx)
		if err == nil {
			settings, err := service.Settings.Get("timezone").Do()
			if err == nil && settings != nil {
				// Email not directly available from settings, but we know auth works
				email = "authenticated"
			}
		}
	}

	return true, email, token.Expiry, nil
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
