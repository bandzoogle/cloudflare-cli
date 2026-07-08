package cmd

import (
	"fmt"
	"time"

	"github.com/bandzoogle/cloudflare-cli/internal/output"
	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
)

var zonesBotManagementCmd = &cobra.Command{
	Use:   "bot-management",
	Short: "Bot Management settings (read-only)",
}

var zonesBotManagementGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get Bot Management configuration for a zone",
	Long: `Read Bot Management / AI bot protection settings for a zone.

Includes ai_bots_protection, Super Bot Fight Mode heuristics, and related flags.`,
	Example: `  cfcli zones bot-management get --zone-name bandzoogle.com`,
	PreRunE: validateZoneFlags,
	RunE:    runZonesBotManagementGet,
}

var zonesRulesetsCmd = &cobra.Command{
	Use:   "rulesets",
	Short: "Zone rulesets (read-only)",
}

type rulesetSummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Phase         string `json:"phase"`
	Kind          string `json:"kind"`
	Rules         int    `json:"rules"`
	EnabledRules  int    `json:"enabled_rules"`
}

var zonesRulesetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rulesets deployed on a zone",
	Example: `  cfcli zones rulesets list --zone-name bandzoogle.com`,
	PreRunE: validateZoneFlags,
	RunE:    runZonesRulesetsList,
}

var zonesRulesetsGetCmd = &cobra.Command{
	Use:   "get RULESET_ID",
	Short: "Get a ruleset by ID (includes rules)",
	Example: `  cfcli zones rulesets get efb7b8c949ac4650a09736fc376e9aee --zone-name bandzoogle.com`,
	Args:  cobra.ExactArgs(1),
	PreRunE: validateZoneFlags,
	RunE:    runZonesRulesetsGet,
}

var zonesWafCmd = &cobra.Command{
	Use:   "waf",
	Short: "WAF analytics (read-only)",
}

var zonesWafAnalyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Summarize recent firewall block events for a zone",
	Long: `Query Cloudflare GraphQL firewallEventsAdaptiveGroups for a zone.

Groups blocked/challenged requests by rule description and host. Requires
analytics read permission on the API token.`,
	Example: `  cfcli zones waf analytics --zone-name bandzoogle.com --since 7d
  cfcli zones waf analytics --zone-id ZONE_ID --since 24h --limit 15`,
	PreRunE: validateZoneFlags,
	RunE:    runZonesWAFAnalytics,
}

func init() {
	zonesCmd.AddCommand(zonesBotManagementCmd)
	zonesBotManagementCmd.AddCommand(zonesBotManagementGetCmd)

	zonesCmd.AddCommand(zonesRulesetsCmd)
	zonesRulesetsCmd.AddCommand(zonesRulesetsListCmd)
	zonesRulesetsCmd.AddCommand(zonesRulesetsGetCmd)

	zonesCmd.AddCommand(zonesWafCmd)
	zonesWafCmd.AddCommand(zonesWafAnalyticsCmd)

	for _, c := range []*cobra.Command{
		zonesBotManagementGetCmd,
		zonesRulesetsListCmd,
		zonesRulesetsGetCmd,
		zonesWafAnalyticsCmd,
	} {
		c.Flags().String("zone-id", "", "Zone ID")
		c.Flags().String("zone-name", "", "Zone apex domain name (resolved to zone ID)")
	}

	zonesWafAnalyticsCmd.Flags().String("since", "24h", "Lookback window (e.g. 24h, 72h, 7d)")
	zonesWafAnalyticsCmd.Flags().Int("limit", 25, "Maximum grouped rows to return")
}

func runZonesBotManagementGet(cmd *cobra.Command, args []string) error {
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

	bm, err := client.API.GetBotManagement(ctx, cloudflare.ZoneIdentifier(zoneID))
	if err != nil {
		return apiError("zones bot-management get", err)
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "zones bot-management get", "zone_id": zoneID},
		meta(client.DefaultAccountID(), map[string]any{"zone_query": zname}),
		bm,
		outputOptions(),
	)
}

func runZonesRulesetsList(cmd *cobra.Command, args []string) error {
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

	rulesets, err := client.API.ListRulesets(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListRulesetsParams{})
	if err != nil {
		return apiError("zones rulesets list", err)
	}

	summaries := make([]rulesetSummary, 0, len(rulesets))
	for _, rs := range rulesets {
		summaries = append(summaries, summarizeRuleset(rs))
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "zones rulesets list", "zone_id": zoneID},
		meta(client.DefaultAccountID(), map[string]any{
			"zone_query": zname,
			"count":      len(summaries),
		}),
		summaries,
		outputOptions(),
	)
}

func runZonesRulesetsGet(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	ctx := commandContext()
	zid, _ := cmd.Flags().GetString("zone-id")
	zname, _ := cmd.Flags().GetString("zone-name")
	rulesetID := args[0]
	zoneID, err := resolveZoneFlags(ctx, client, zid, zname)
	if err != nil {
		return err
	}

	rs, err := client.API.GetRuleset(ctx, cloudflare.ZoneIdentifier(zoneID), rulesetID)
	if err != nil {
		return apiError("zones rulesets get", err)
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "zones rulesets get", "zone_id": zoneID, "ruleset_id": rulesetID},
		meta(client.DefaultAccountID(), map[string]any{
			"zone_query":    zname,
			"enabled_rules": countEnabledRules(rs.Rules),
			"rules":         len(rs.Rules),
		}),
		rs,
		outputOptions(),
	)
}

func runZonesWAFAnalytics(cmd *cobra.Command, args []string) error {
	client, err := cfClient(cmd)
	if err != nil {
		return err
	}
	ctx := commandContext()
	zid, _ := cmd.Flags().GetString("zone-id")
	zname, _ := cmd.Flags().GetString("zone-name")
	sinceFlag, _ := cmd.Flags().GetString("since")
	limit, _ := cmd.Flags().GetInt("limit")

	lookback, err := parseSinceDuration(sinceFlag)
	if err != nil {
		return err
	}
	zoneID, err := resolveZoneFlags(ctx, client, zid, zname)
	if err != nil {
		return err
	}

	groups, err := client.ListFirewallEventGroups(ctx, zoneID, time.Now().Add(-lookback), limit)
	if err != nil {
		return apiError("zones waf analytics", err)
	}

	total := 0
	for _, g := range groups {
		total += g.Count
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "zones waf analytics", "zone_id": zoneID, "since": sinceFlag},
		meta(client.DefaultAccountID(), map[string]any{
			"zone_query":   zname,
			"count":        len(groups),
			"total_events": total,
		}),
		groups,
		outputOptions(),
	)
}

func summarizeRuleset(rs cloudflare.Ruleset) rulesetSummary {
	return rulesetSummary{
		ID:           rs.ID,
		Name:         rs.Name,
		Phase:        rs.Phase,
		Kind:         rs.Kind,
		Rules:        len(rs.Rules),
		EnabledRules: countEnabledRules(rs.Rules),
	}
}

func countEnabledRules(rules []cloudflare.RulesetRule) int {
	n := 0
	for _, rule := range rules {
		if rule.Enabled != nil && *rule.Enabled {
			n++
		}
	}
	return n
}

func parseSinceDuration(raw string) (time.Duration, error) {
	raw = trimSpace(raw)
	if raw == "" {
		return 24 * time.Hour, nil
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d, nil
	}
	if len(raw) >= 2 && raw[len(raw)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(raw, "%dd", &days); err != nil {
			return 0, fmt.Errorf("invalid --since %q: use Go duration (24h) or days (7d)", raw)
		}
		if days <= 0 {
			return 0, fmt.Errorf("--since days must be positive")
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return 0, fmt.Errorf("invalid --since %q: use Go duration (24h) or days (7d)", raw)
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
