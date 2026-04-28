package cmd

import (
	"github.com/bandzoogle/cloudflare-cli/internal/output"
	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
)

var zonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "List and inspect zones",
}

var zonesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List zones accessible to the token",
	Example: `  cfcli zones list
  cfcli zones list --name example.com`,
	RunE: runZonesList,
}

var zonesGetCmd = &cobra.Command{
	Use:   "get ZONE_ID",
	Short: "Get zone details by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runZonesGet,
}

func init() {
	rootCmd.AddCommand(zonesCmd)
	zonesCmd.AddCommand(zonesListCmd)
	zonesCmd.AddCommand(zonesGetCmd)

	zonesListCmd.Flags().String("name", "", "Filter by zone name (exact)")
}

func runZonesList(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	name, _ := cmd.Flags().GetString("name")

	ctx := commandContext()
	var zones []cloudflare.Zone
	var listErr error
	if name != "" {
		zones, listErr = client.API.ListZones(ctx, name)
	} else {
		var zr cloudflare.ZonesResponse
		zr, listErr = client.API.ListZonesContext(ctx)
		zones = zr.Result
	}
	if listErr != nil {
		return apiError("zones list", listErr)
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "zones list", "name": name},
		meta(client.DefaultAccountID(), map[string]any{"count": len(zones)}),
		zones,
		outputOptions(),
	)
}

func runZonesGet(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	zoneID := args[0]

	z, err := client.API.ZoneDetails(commandContext(), zoneID)
	if err != nil {
		return apiError("zones get", err)
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "zones get", "zone_id": zoneID},
		meta(client.DefaultAccountID(), nil),
		z,
		outputOptions(),
	)
}
