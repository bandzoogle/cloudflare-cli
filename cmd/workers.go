package cmd

import (
	"github.com/bandzoogle/cloudflare-cli/internal/output"
	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
)

var workersCmd = &cobra.Command{
	Use:   "workers",
	Short: "Workers scripts (read-only listing)",
}

var workersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Workers scripts for an account",
	Long: `Requires an account ID via --account-id or CLOUDFLARE_ACCOUNT_ID.`,
	Example: `  cfcli workers list --account-id $CLOUDFLARE_ACCOUNT_ID
  cfcli workers list`,
	RunE: runWorkersList,
}

func init() {
	rootCmd.AddCommand(workersCmd)
	workersCmd.AddCommand(workersListCmd)
}

func runWorkersList(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	accountID, err := client.RequireAccountID()
	if err != nil {
		return err
	}

	resp, ri, err := client.API.ListWorkers(commandContext(), cloudflare.AccountIdentifier(accountID), cloudflare.ListWorkersParams{})
	if err != nil {
		return apiError("workers list", err)
	}

	data := map[string]any{
		"scripts": resp.WorkerList,
	}
	if ri != nil {
		data["result_info"] = ri
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "workers list", "account_id": accountID},
		meta(accountID, map[string]any{}),
		data,
		outputOptions(),
	)
}
