package commands

import (
	"github.com/btafoya/gcal-cli/pkg/output"
	"github.com/btafoya/gcal-cli/pkg/types"
	"github.com/spf13/cobra"
)

var (
	// Version is the application version (set at build time)
	Version = "dev"
	// Commit is the git commit hash (set at build time)
	Commit = "unknown"
	// BuildDate is the build date (set at build time)
	BuildDate = "unknown"
)

// NewVersionCommand creates the version command
func NewVersionCommand(formatter output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print detailed version information including version, commit, and build date",
		Run: func(cmd *cobra.Command, args []string) {
			versionData := map[string]interface{}{
				"version":   Version,
				"commit":    Commit,
				"buildDate": BuildDate,
			}

			response := types.SuccessResponse("version", versionData)
			output, err := formatter.Format(response)
			if err != nil {
				cmd.PrintErrf("Error formatting output: %v\n", err)
				return
			}
			cmd.Println(output)
		},
	}
}
