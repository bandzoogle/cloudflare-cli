package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bandzoogle/cloudflare-cli/internal/output"
	"github.com/spf13/cobra"
)

type scopeInfo struct {
	Command        string   `json:"command"`
	API            string   `json:"api"`
	Permission     string   `json:"permission"`
	TokenTemplates []string `json:"token_templates,omitempty"`
	Notes          []string `json:"notes,omitempty"`
}

type scopesResponse struct {
	Summary         []string    `json:"summary"`
	Required        []scopeInfo `json:"required"`
	Unique          []string    `json:"unique"`
	RecommendedRole []string    `json:"recommended_role"`
}

var scopesCmd = &cobra.Command{
	Use:     "scopes",
	Aliases: []string{"permissions"},
	Short:   "Cloudflare API token permissions used by cfcli",
	Long: `Show the Cloudflare API token permissions needed by cfcli read-only commands.

Cloudflare uses granular permissions on API tokens. Grant only what you need.`,
	RunE: runScopes,
}

func init() {
	rootCmd.AddCommand(scopesCmd)
	scopesCmd.Flags().String("command", "", "Filter by command prefix, e.g. dns, zones, workers")
}

func runScopes(cmd *cobra.Command, args []string) error {
	filter, _ := cmd.Flags().GetString("command")
	required := filterScopes(requiredScopes(), filter)
	resp := scopesResponse{
		Summary: []string{
			"Create a dedicated API token with read-only permissions for the routes you need.",
			"Permissions are assigned in the Cloudflare dashboard under My Profile → API Tokens.",
			"Exact labels match the permission groups shown when editing a custom token.",
		},
		Required: required,
		Unique:   uniquePermissions(required),
		RecommendedRole: []string{
			"Prefer Account → Account API Tokens for automation tied to a single account.",
			"Use zone-level restriction when the token should only see specific zones.",
		},
	}

	opts := outputOptions()
	if !opts.Raw {
		_, err := fmt.Fprint(cmd.OutOrStdout(), renderScopesText(resp, filter))
		return err
	}

	return output.WriteEnvelope(cmd.OutOrStdout(),
		map[string]any{"command": "scopes", "filter": filter},
		map[string]any{"auth_required": false, "count": len(required)},
		resp,
		opts,
	)
}

func renderScopesText(resp scopesResponse, filter string) string {
	var b strings.Builder
	if filter != "" {
		fmt.Fprintf(&b, "Cloudflare token permissions for %q:\n\n", filter)
	} else {
		b.WriteString("Cloudflare token permissions for cfcli:\n\n")
	}

	b.WriteString(strings.Join(resp.Unique, "\n"))
	b.WriteString("\n\n")
	b.WriteString("Command coverage:\n")
	for _, item := range resp.Required {
		fmt.Fprintf(&b, "- %s: %s\n", item.Command, item.Permission)
	}

	b.WriteString("\nNotes:\n")
	for _, note := range resp.RecommendedRole {
		fmt.Fprintf(&b, "- %s\n", note)
	}
	return b.String()
}

func requiredScopes() []scopeInfo {
	return []scopeInfo{
		{
			Command:    "zones list|get",
			API:        "Zones API",
			Permission: "Zone → Zone → Read",
			TokenTemplates: []string{
				"Read all resources",
				"Custom: Zone / Zone / Read",
			},
		},
		{
			Command:    "zones settings list|get",
			API:        "Zone Settings API",
			Permission: "Zone → Zone Settings → Read",
			TokenTemplates: []string{
				"Custom: Zone / Zone Settings / Read",
			},
		},
		{
			Command:    "dns list|get",
			API:        "DNS Records API",
			Permission: "Zone → DNS → Read",
			TokenTemplates: []string{
				"Custom: Zone / DNS / Read",
			},
		},
		{
			Command:    "accounts list|get",
			API:        "Accounts API",
			Permission: "Account → Account Settings → Read",
			TokenTemplates: []string{
				"Read all resources",
				"Custom: Account / Account Settings / Read",
			},
		},
		{
			Command:    "workers list",
			API:        "Workers Scripts API",
			Permission: "Account → Workers Scripts → Read",
			Notes: []string{
				"Requires CLOUDFLARE_ACCOUNT_ID or --account-id in addition to token scopes.",
			},
			TokenTemplates: []string{
				"Custom: Account / Workers Scripts / Read",
			},
		},
		{
			Command:    "user whoami",
			API:        "User API",
			Permission: "User → User Details → Read",
			TokenTemplates: []string{
				"User → Read user details",
				"Custom: User / User Details / Read",
			},
		},
	}
}

func filterScopes(scopes []scopeInfo, filter string) []scopeInfo {
	filter = strings.ToLower(strings.TrimSpace(filter))
	if filter == "" {
		return scopes
	}
	out := make([]scopeInfo, 0, len(scopes))
	for _, scope := range scopes {
		lower := strings.ToLower(scope.Command)
		if strings.HasPrefix(lower, filter) || strings.Contains(lower, " "+filter) {
			out = append(out, scope)
		}
	}
	return out
}

func uniquePermissions(scopes []scopeInfo) []string {
	seen := map[string]bool{}
	for _, scope := range scopes {
		if scope.Permission != "" {
			seen[scope.Permission] = true
		}
	}
	out := make([]string, 0, len(seen))
	for permission := range seen {
		out = append(out, permission)
	}
	sort.Strings(out)
	return out
}
