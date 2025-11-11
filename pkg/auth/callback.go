package auth

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/btafoya/gcal-cli/pkg/types"
)

const (
	// CallbackPath is the OAuth callback path
	CallbackPath = "/oauth/callback"
	// DefaultCallbackPort is the default port for the callback server
	DefaultCallbackPort = 8080
)

// CallbackServer handles OAuth2 callbacks
type CallbackServer struct {
	Port         int
	State        string
	CodeChan     chan string
	ErrorChan    chan error
	Server       *http.Server
	ShutdownChan chan struct{}
}

// NewCallbackServer creates a new callback server
func NewCallbackServer(port int, state string) *CallbackServer {
	return &CallbackServer{
		Port:         port,
		State:        state,
		CodeChan:     make(chan string, 1),
		ErrorChan:    make(chan error, 1),
		ShutdownChan: make(chan struct{}, 1),
	}
}

// Start starts the callback server
func (cs *CallbackServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(CallbackPath, cs.handleCallback)

	cs.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cs.Port),
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		if err := cs.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cs.ErrorChan <- types.ErrNetworkError("callback server failed").
				WithWrappedError(err)
		}
	}()

	return nil
}

// WaitForCode waits for the authorization code or error
func (cs *CallbackServer) WaitForCode(timeout time.Duration) (code string, err error) {
	select {
	case code = <-cs.CodeChan:
		return code, nil
	case err = <-cs.ErrorChan:
		return "", err
	case <-time.After(timeout):
		return "", types.ErrAuthFailed("authentication timeout").
			WithDetails(fmt.Sprintf("no response received after %v", timeout)).
			WithSuggestedAction("Try the authentication flow again")
	}
}

// Shutdown gracefully shuts down the callback server
func (cs *CallbackServer) Shutdown(ctx context.Context) error {
	if cs.Server != nil {
		if err := cs.Server.Shutdown(ctx); err != nil {
			return types.ErrNetworkError("failed to shutdown callback server").
				WithWrappedError(err)
		}
	}
	close(cs.ShutdownChan)
	return nil
}

// handleCallback handles the OAuth2 callback request
func (cs *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	query := r.URL.Query()
	code := query.Get("code")
	state := query.Get("state")
	errorParam := query.Get("error")
	errorDesc := query.Get("error_description")

	// Check for errors
	if errorParam != "" {
		errMsg := fmt.Sprintf("OAuth error: %s", errorParam)
		if errorDesc != "" {
			errMsg = fmt.Sprintf("%s (%s)", errMsg, errorDesc)
		}
		cs.ErrorChan <- types.ErrAuthFailed(errMsg)
		cs.renderErrorPage(w, errMsg)
		return
	}

	// Validate state for CSRF protection
	if state != cs.State {
		err := types.ErrAuthFailed("invalid state parameter").
			WithDetails("possible CSRF attack").
			WithSuggestedAction("Try the authentication flow again")
		cs.ErrorChan <- err
		cs.renderErrorPage(w, "Authentication failed: invalid state parameter")
		return
	}

	// Validate code
	if code == "" {
		err := types.ErrAuthFailed("no authorization code received")
		cs.ErrorChan <- err
		cs.renderErrorPage(w, "Authentication failed: no authorization code")
		return
	}

	// Send code to channel
	cs.CodeChan <- code

	// Render success page
	cs.renderSuccessPage(w)
}

// renderSuccessPage renders a success page
func (cs *CallbackServer) renderSuccessPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	tmpl := template.Must(template.New("success").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f5f5f5;
        }
        .container {
            text-align: center;
            padding: 40px;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            max-width: 500px;
        }
        .success-icon {
            font-size: 64px;
            color: #4CAF50;
        }
        h1 {
            color: #333;
            margin: 20px 0;
        }
        p {
            color: #666;
            line-height: 1.6;
        }
        .close-message {
            margin-top: 20px;
            padding: 10px;
            background-color: #f0f0f0;
            border-radius: 4px;
            font-size: 14px;
            color: #555;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">✓</div>
        <h1>Authentication Successful!</h1>
        <p>You have successfully authenticated with Google Calendar.</p>
        <p>You can now close this window and return to the terminal.</p>
        <div class="close-message">
            This window will close automatically, or you can close it manually.
        </div>
    </div>
    <script>
        // Auto-close window after 3 seconds
        setTimeout(function() {
            window.close();
        }, 3000);
    </script>
</body>
</html>
`))

	tmpl.Execute(w, nil)
}

// renderErrorPage renders an error page
func (cs *CallbackServer) renderErrorPage(w http.ResponseWriter, errorMessage string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)

	tmpl := template.Must(template.New("error").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Failed</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f5f5f5;
        }
        .container {
            text-align: center;
            padding: 40px;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            max-width: 500px;
        }
        .error-icon {
            font-size: 64px;
            color: #f44336;
        }
        h1 {
            color: #333;
            margin: 20px 0;
        }
        p {
            color: #666;
            line-height: 1.6;
        }
        .error-details {
            margin-top: 20px;
            padding: 15px;
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            text-align: left;
            font-size: 14px;
            color: #856404;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="error-icon">✗</div>
        <h1>Authentication Failed</h1>
        <p>There was a problem authenticating with Google Calendar.</p>
        <div class="error-details">
            <strong>Error:</strong> {{.}}
        </div>
        <p style="margin-top: 20px;">Please return to the terminal and try again.</p>
    </div>
</body>
</html>
`))

	tmpl.Execute(w, errorMessage)
}

// GetCallbackURL returns the full callback URL for this server
func (cs *CallbackServer) GetCallbackURL() string {
	return fmt.Sprintf("http://localhost:%d%s", cs.Port, CallbackPath)
}
