package cmd

import (
	"github.com/bandzoogle/cloudflare-cli/internal/output"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Current user / token context",
}

var userWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Return user details for the API token (token smoke test)",
	RunE:  runUserWhoami,
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userWhoamiCmd)
}

func runUserWhoami(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}

	u, err := client.API.UserDetails(commandContext())
	if err != nil {
		return apiError("user whoami", err)
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "user whoami"},
		meta(client.DefaultAccountID(), nil),
		u,
		outputOptions(),
	)
}
