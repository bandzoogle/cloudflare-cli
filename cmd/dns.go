package cmd

import (
	"context"
	"fmt"

	"github.com/bandzoogle/cloudflare-cli/internal/cf"
	"github.com/bandzoogle/cloudflare-cli/internal/output"
	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS records (read-only)",
}

var dnsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List DNS records for a zone",
	Long: `List DNS records for a zone. Provide --zone-id or --zone-name (domain).

Zone names are resolved via the Zones API; ambiguous matches error out.`,
	Example: `  cfcli dns list --zone-id 7c5dae5552338874cf36954f019713fa
  cfcli dns list --zone-name example.com --type A`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		zid, _ := cmd.Flags().GetString("zone-id")
		zname, _ := cmd.Flags().GetString("zone-name")
		if zid == "" && zname == "" {
			return fmt.Errorf("one of --zone-id or --zone-name is required")
		}
		if zid != "" && zname != "" {
			return fmt.Errorf("pass only one of --zone-id or --zone-name")
		}
		return nil
	},
	RunE: runDNSList,
}

var dnsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a DNS record by ID",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := requireStringFlag(cmd, "record-id"); err != nil {
			return err
		}
		zid, _ := cmd.Flags().GetString("zone-id")
		zname, _ := cmd.Flags().GetString("zone-name")
		if zid == "" && zname == "" {
			return fmt.Errorf("--zone-id or --zone-name is required")
		}
		if zid != "" && zname != "" {
			return fmt.Errorf("pass only one of --zone-id or --zone-name")
		}
		return nil
	},
	RunE: runDNSGet,
}

func init() {
	rootCmd.AddCommand(dnsCmd)
	dnsCmd.AddCommand(dnsListCmd)
	dnsCmd.AddCommand(dnsGetCmd)

	for _, c := range []*cobra.Command{dnsListCmd, dnsGetCmd} {
		c.Flags().String("zone-id", "", "Zone ID")
		c.Flags().String("zone-name", "", "Zone apex domain name (resolved to zone ID)")
	}
	dnsListCmd.Flags().String("type", "", "DNS record type filter (e.g. A, CNAME)")
	dnsGetCmd.Flags().String("record-id", "", "DNS record ID")
}

func resolveDNSZoneID(ctx context.Context, client *cf.Client, zoneID, zoneName string) (string, error) {
	switch {
	case zoneID != "":
		return zoneID, nil
	case zoneName != "":
		return client.ResolveZoneID(ctx, zoneName)
	default:
		return "", fmt.Errorf("zone not specified")
	}
}

func runDNSList(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	ctx := commandContext()
	zid, _ := cmd.Flags().GetString("zone-id")
	zname, _ := cmd.Flags().GetString("zone-name")
	zoneID, err := resolveDNSZoneID(ctx, client, zid, zname)
	if err != nil {
		return err
	}

	recType, _ := cmd.Flags().GetString("type")
	params := cloudflare.ListDNSRecordsParams{Type: recType}

	records, ri, err := client.API.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), params)
	if err != nil {
		return apiError("dns list", err)
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "dns list", "zone_id": zoneID, "type": recType},
		meta(client.DefaultAccountID(), map[string]any{
			"zone_query": zname,
			"page":       ri.Page,
			"total":      ri.Total,
		}),
		records,
		outputOptions(),
	)
}

func runDNSGet(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	ctx := commandContext()
	zid, _ := cmd.Flags().GetString("zone-id")
	zname, _ := cmd.Flags().GetString("zone-name")
	recordID, _ := cmd.Flags().GetString("record-id")

	zoneID, err := resolveDNSZoneID(ctx, client, zid, zname)
	if err != nil {
		return err
	}

	rec, err := client.API.GetDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), recordID)
	if err != nil {
		return apiError("dns get", err)
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "dns get", "zone_id": zoneID, "record_id": recordID},
		meta(client.DefaultAccountID(), map[string]any{"zone_query": zname}),
		rec,
		outputOptions(),
	)
}
