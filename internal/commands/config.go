package commands

import (
	"fmt"

	"github.com/btafoya/gcal-cli/pkg/config"
	"github.com/btafoya/gcal-cli/pkg/output"
	"github.com/btafoya/gcal-cli/pkg/types"
	"github.com/spf13/cobra"
)

// NewConfigCommand creates the config command group
func NewConfigCommand(formatter output.Formatter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Manage gcal-cli configuration settings",
	}

	cmd.AddCommand(newConfigShowCommand(formatter))
	cmd.AddCommand(newConfigInitCommand(formatter))
	cmd.AddCommand(newConfigSetCommand(formatter))

	return cmd
}

func newConfigShowCommand(formatter output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		Long:  "Display the current configuration settings",
		Run: func(cmd *cobra.Command, args []string) {
			configStr, err := config.DisplayConfig()
			if err != nil {
				appErr := types.ErrConfigError("failed to display configuration").
					WithWrappedError(err)
				response := types.ErrorResponse(appErr)
				output, _ := formatter.Format(response)
				cmd.Println(output)
				return
			}

			// For text format, output directly
			if _, ok := formatter.(*output.TextFormatter); ok {
				cmd.Println(configStr)
				return
			}

			// For other formats, use structured response
			response := types.SuccessResponse("config_show", map[string]interface{}{
				"config": configStr,
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

func newConfigInitCommand(formatter output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Long:  "Initialize configuration directory and create default config file",
		Run: func(cmd *cobra.Command, args []string) {
			configDir, err := config.EnsureConfigDir()
			if err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrConfigError("failed to create config directory").
						WithWrappedError(err)
				}
				response := types.ErrorResponse(appErr)
				output, _ := formatter.Format(response)
				cmd.Println(output)
				return
			}

			if err := config.Save(); err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrConfigError("failed to save config file").
						WithWrappedError(err)
				}
				response := types.ErrorResponse(appErr)
				output, _ := formatter.Format(response)
				cmd.Println(output)
				return
			}

			response := types.SuccessResponse("config_init", map[string]interface{}{
				"message":    "Configuration initialized successfully",
				"configDir":  configDir,
				"configFile": fmt.Sprintf("%s/config.yaml", configDir),
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

func newConfigSetCommand(formatter output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  "Set a configuration value and save to config file",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]

			config.Set(key, value)

			if err := config.Save(); err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrConfigError("failed to save config file").
						WithWrappedError(err)
				}
				response := types.ErrorResponse(appErr)
				output, _ := formatter.Format(response)
				cmd.Println(output)
				return
			}

			response := types.SuccessResponse("config_set", map[string]interface{}{
				"message": "Configuration value set successfully",
				"key":     key,
				"value":   value,
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
