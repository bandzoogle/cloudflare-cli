package cmd

import (
	"strings"
	"testing"
)

func TestUniquePermissionsDedupesAndSorts(t *testing.T) {
	got := uniquePermissions([]scopeInfo{
		{Permission: "Zone → DNS → Read"},
		{Permission: "Zone → Zone → Read"},
		{Permission: "Zone → DNS → Read"},
	})
	want := []string{"Zone → DNS → Read", "Zone → Zone → Read"}
	if len(got) != len(want) {
		t.Fatalf("expected %d permissions, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %#v, got %#v", want, got)
		}
	}
}

func TestFilterScopesByCommandPrefix(t *testing.T) {
	got := filterScopes(requiredScopes(), "dns")
	if len(got) != 1 {
		t.Fatalf("expected one dns scope row, got %d", len(got))
	}
	if !strings.Contains(got[0].Command, "dns") {
		t.Fatalf("unexpected row: %s", got[0].Command)
	}
}

func TestRenderScopesTextShowsCommandCoverage(t *testing.T) {
	resp := scopesResponse{
		Required: []scopeInfo{
			{Command: "zones list", Permission: "Zone → Zone → Read"},
		},
		Unique: []string{"Zone → Zone → Read"},
		RecommendedRole: []string{
			"Prefer account API tokens for automation.",
		},
	}

	got := renderScopesText(resp, "")
	if !strings.Contains(got, "- zones list: Zone → Zone → Read") {
		t.Fatalf("expected command coverage, got:\n%s", got)
	}
}
