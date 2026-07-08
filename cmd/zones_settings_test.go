package cmd

import (
	"testing"
)

func TestValidateZoneFlagsRequiresZone(t *testing.T) {
	cmd := zonesSettingsListCmd
	_ = cmd.Flags().Set("zone-id", "")
	_ = cmd.Flags().Set("zone-name", "")
	cmd.SetArgs([]string{})
	if err := validateZoneFlags(cmd, nil); err == nil {
		t.Fatal("expected error when zone flags missing")
	}
}

func TestValidateZoneFlagsRejectsBothZoneSelectors(t *testing.T) {
	cmd := zonesSettingsListCmd
	_ = cmd.Flags().Set("zone-id", "abc")
	_ = cmd.Flags().Set("zone-name", "example.com")
	if err := validateZoneFlags(cmd, nil); err == nil {
		t.Fatal("expected error when both zone-id and zone-name set")
	}
	_ = cmd.Flags().Set("zone-id", "")
	_ = cmd.Flags().Set("zone-name", "")
}

func TestRequiredScopesIncludesZoneSettings(t *testing.T) {
	got := filterScopes(requiredScopes(), "zones settings")
	if len(got) != 1 {
		t.Fatalf("expected one zones settings scope row, got %d", len(got))
	}
	if got[0].Permission != "Zone → Zone Settings → Read" {
		t.Fatalf("unexpected permission: %s", got[0].Permission)
	}
}
