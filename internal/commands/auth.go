package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/btafoya/gcal-cli/pkg/auth"
	"github.com/btafoya/gcal-cli/pkg/config"
	"github.com/btafoya/gcal-cli/pkg/examples"
	"github.com/btafoya/gcal-cli/pkg/output"
	"github.com/btafoya/gcal-cli/pkg/types"
	"github.com/spf13/cobra"
)

// NewAuthCommand creates the auth command group
func NewAuthCommand(formatter output.Formatter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Manage Google Calendar authentication credentials and status",
	}

	cmd.AddCommand(newAuthLoginCommand(formatter))
	cmd.AddCommand(newAuthLogoutCommand(formatter))
	cmd.AddCommand(newAuthStatusCommand(formatter))

	return cmd
}

func newAuthLoginCommand(formatter output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:     "login",
		Short:   "Authenticate with Google Calendar",
		Long:    "Start the OAuth2 authentication flow to obtain and store Google Calendar credentials",
		Example: examples.AuthLoginExamples,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Get credentials and token paths from config
			credentialsPath := config.GetString("auth.credentials_path")
			tokenPath := config.GetString("auth.token_path")

			// Create auth manager
			manager, err := auth.NewManager(credentialsPath, tokenPath)
			if err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrAuthFailed("authentication initialization failed").
						WithWrappedError(err)
				}
				response := types.ErrorResponse(appErr)
				output, _ := formatter.Format(response)
				cmd.Println(output)
				return
			}

			// Perform login
			token, err := manager.Login(ctx)
			if err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrAuthFailed("authentication failed").
						WithWrappedError(err)
				}
				response := types.ErrorResponse(appErr)
				output, _ := formatter.Format(response)
				cmd.Println(output)
				return
			}

			// Get user info
			email, _ := auth.GetUserInfo(token)
			if email == "" {
				email = "authenticated"
			}

			response := types.SuccessResponse("auth_login", map[string]interface{}{
				"message":    "Successfully authenticated with Google Calendar",
				"email":      email,
				"expires_at": token.Expiry.Format(time.RFC3339),
			})
			output, err := formatter.Format(response)
			if err != nil {
				cmd.PrintErrf("Error formatting output: %v\n", err)
				return
			}
			cmd.Println(output)
		},
	}
}

func newAuthLogoutCommand(formatter output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:     "logout",
		Short:   "Remove authentication credentials",
		Long:    "Delete stored authentication token and remove Google Calendar access",
		Example: examples.AuthLogoutExamples,
		Run: func(cmd *cobra.Command, args []string) {
			// Get token path from config
			tokenPath := config.GetString("auth.token_path")

			// Create auth manager
			manager, err := auth.NewManager(
				config.GetString("auth.credentials_path"),
				tokenPath,
			)
			if err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrAuthFailed("logout initialization failed").
						WithWrappedError(err)
				}
				response := types.ErrorResponse(appErr)
				output, _ := formatter.Format(response)
				cmd.Println(output)
				return
			}

			// Perform logout
			if err := manager.Logout(); err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrAuthFailed("logout failed").
						WithWrappedError(err)
				}
				response := types.ErrorResponse(appErr)
				output, _ := formatter.Format(response)
				cmd.Println(output)
				return
			}

			response := types.SuccessResponse("auth_logout", map[string]interface{}{
				"message": "Successfully logged out and removed authentication credentials",
			})
			output, err := formatter.Format(response)
			if err != nil {
				cmd.PrintErrf("Error formatting output: %v\n", err)
				return
			}
			cmd.Println(output)
		},
	}
}

func newAuthStatusCommand(formatter output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Check authentication status",
		Long:    "Display current authentication status including token expiry and user email",
		Example: examples.AuthStatusExamples,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Get credentials and token paths from config
			credentialsPath := config.GetString("auth.credentials_path")
			tokenPath := config.GetString("auth.token_path")

			// Create auth manager
			manager, err := auth.NewManager(credentialsPath, tokenPath)
			if err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrAuthFailed("status check initialization failed").
						WithWrappedError(err)
				}
				response := types.ErrorResponse(appErr)
				output, _ := formatter.Format(response)
				cmd.Println(output)
				return
			}

			// Check authentication status
			authenticated, email, expiresAt, err := manager.CheckAuthStatus(ctx)
			if err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrAuthFailed("status check failed").
						WithWrappedError(err)
				}
				response := types.ErrorResponse(appErr)
				output, _ := formatter.Format(response)
				cmd.Println(output)
				return
			}

			// Build status response
			statusData := map[string]interface{}{
				"authenticated": authenticated,
			}

			if authenticated {
				statusData["email"] = email
				statusData["expires_at"] = expiresAt.Format(time.RFC3339)

				// Calculate time until expiration
				if !expiresAt.IsZero() {
					timeUntilExpiry := time.Until(expiresAt)
					if timeUntilExpiry > 0 {
						statusData["expires_in"] = timeUntilExpiry.String()
					} else {
						statusData["expired"] = true
					}
				}

				statusData["message"] = fmt.Sprintf("Authenticated as %s", email)
			} else {
				statusData["message"] = "Not authenticated. Run 'gcal-cli auth login' to authenticate."
			}

			response := types.SuccessResponse("auth_status", statusData)
			output, err := formatter.Format(response)
			if err != nil {
				cmd.PrintErrf("Error formatting output: %v\n", err)
				return
			}
			cmd.Println(output)
		},
	}
}
