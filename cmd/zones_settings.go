package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/bandzoogle/cloudflare-cli/internal/cf"
	"github.com/bandzoogle/cloudflare-cli/internal/output"
	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
)

var zonesSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Zone settings (read-only)",
}

var zonesSettingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List zone settings",
	Long: `List all zone settings for a zone. Provide --zone-id or --zone-name (domain).

Useful for inspecting proxy-related toggles such as email_obfuscation, rocket_loader, ssl, and minify.`,
	Example: `  cfcli zones settings list --zone-name bandzoogle.com
  cfcli zones settings list --zone-id 23d6a9d818228c7296277cdacd91df49`,
	PreRunE: validateZoneFlags,
	RunE:    runZonesSettingsList,
}

var zonesSettingsGetCmd = &cobra.Command{
	Use:   "get SETTING",
	Short: "Get a single zone setting",
	Long: `Get one zone setting by ID (for example email_obfuscation, ssl, rocket_loader).

Provide --zone-id or --zone-name (domain).`,
	Example: `  cfcli zones settings get email_obfuscation --zone-name bandzoogle.com
  cfcli zones settings get ssl --zone-id 23d6a9d818228c7296277cdacd91df49`,
	Args:    cobra.ExactArgs(1),
	PreRunE: validateZoneFlags,
	RunE:    runZonesSettingsGet,
}

func init() {
	zonesCmd.AddCommand(zonesSettingsCmd)
	zonesSettingsCmd.AddCommand(zonesSettingsListCmd)
	zonesSettingsCmd.AddCommand(zonesSettingsGetCmd)

	for _, c := range []*cobra.Command{zonesSettingsListCmd, zonesSettingsGetCmd} {
		c.Flags().String("zone-id", "", "Zone ID")
		c.Flags().String("zone-name", "", "Zone apex domain name (resolved to zone ID)")
	}
}

func validateZoneFlags(cmd *cobra.Command, args []string) error {
	zid, _ := cmd.Flags().GetString("zone-id")
	zname, _ := cmd.Flags().GetString("zone-name")
	if zid == "" && zname == "" {
		return fmt.Errorf("one of --zone-id or --zone-name is required")
	}
	if zid != "" && zname != "" {
		return fmt.Errorf("pass only one of --zone-id or --zone-name")
	}
	return nil
}

func resolveZoneFlags(ctx context.Context, client *cf.Client, zoneID, zoneName string) (string, error) {
	return resolveDNSZoneID(ctx, client, zoneID, zoneName)
}

func runZonesSettingsList(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	ctx := commandContext()
	zid, _ := cmd.Flags().GetString("zone-id")
	zname, _ := cmd.Flags().GetString("zone-name")
	zoneID, err := resolveZoneFlags(ctx, client, zid, zname)
	if err != nil {
		return err
	}

	resp, err := client.API.ZoneSettings(ctx, zoneID)
	if err != nil {
		return apiError("zones settings list", err)
	}

	settings := resp.Result
	if settings == nil {
		settings = []cloudflare.ZoneSetting{}
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "zones settings list", "zone_id": zoneID},
		meta(client.DefaultAccountID(), map[string]any{
			"zone_query": zname,
			"count":      len(settings),
		}),
		settings,
		outputOptions(),
	)
}

func runZonesSettingsGet(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	ctx := commandContext()
	zid, _ := cmd.Flags().GetString("zone-id")
	zname, _ := cmd.Flags().GetString("zone-name")
	settingID := strings.TrimSpace(args[0])
	if settingID == "" {
		return fmt.Errorf("setting ID is required")
	}

	zoneID, err := resolveZoneFlags(ctx, client, zid, zname)
	if err != nil {
		return err
	}

	setting, err := client.API.GetZoneSetting(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.GetZoneSettingParams{
		Name: settingID,
	})
	if err != nil {
		return apiError("zones settings get", err)
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "zones settings get", "zone_id": zoneID, "setting": settingID},
		meta(client.DefaultAccountID(), map[string]any{"zone_query": zname}),
		setting,
		outputOptions(),
	)
}
