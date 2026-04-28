package cmd

import (
	"github.com/bandzoogle/cloudflare-cli/internal/output"
	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Cloudflare accounts",
}

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List accounts the token can access",
	Example: `  cfcli accounts list
  cfcli accounts list --name prod`,
	RunE: runAccountsList,
}

var accountsGetCmd = &cobra.Command{
	Use:   "get ACCOUNT_ID",
	Short: "Get account details",
	Args:  cobra.ExactArgs(1),
	RunE:  runAccountsGet,
}

func init() {
	rootCmd.AddCommand(accountsCmd)
	accountsCmd.AddCommand(accountsListCmd)
	accountsCmd.AddCommand(accountsGetCmd)

	accountsListCmd.Flags().String("name", "", "Filter by account name (substring match via API)")
}

func runAccountsList(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	name, _ := cmd.Flags().GetString("name")
	ctx := commandContext()

	var all []cloudflare.Account
	page := 1
	const perPage = 50
	for {
		accts, ri, err := client.API.Accounts(ctx, cloudflare.AccountsListParams{
			Name: name,
			PaginationOptions: cloudflare.PaginationOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		if err != nil {
			return apiError("accounts list", err)
		}
		all = append(all, accts...)
		if page >= ri.TotalPages || ri.TotalPages == 0 {
			break
		}
		page++
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "accounts list", "name": name},
		meta(client.DefaultAccountID(), map[string]any{"count": len(all)}),
		all,
		outputOptions(),
	)
}

func runAccountsGet(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	accountID := args[0]

	a, _, err := client.API.Account(commandContext(), accountID)
	if err != nil {
		return apiError("accounts get", err)
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "accounts get", "account_id": accountID},
		meta(client.DefaultAccountID(), nil),
		a,
		outputOptions(),
	)
}
