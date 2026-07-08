package cmd

import (
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

func TestParseSinceDurationHours(t *testing.T) {
	d, err := parseSinceDuration("24h")
	if err != nil {
		t.Fatal(err)
	}
	if d != 24*time.Hour {
		t.Fatalf("got %v", d)
	}
}

func TestParseSinceDurationDays(t *testing.T) {
	d, err := parseSinceDuration("7d")
	if err != nil {
		t.Fatal(err)
	}
	if d != 7*24*time.Hour {
		t.Fatalf("got %v", d)
	}
}

func TestParseSinceDurationInvalid(t *testing.T) {
	if _, err := parseSinceDuration("bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCountEnabledRules(t *testing.T) {
	enabled := true
	disabled := false
	got := countEnabledRules([]cloudflare.RulesetRule{
		{Enabled: &enabled},
		{Enabled: &disabled},
		{Enabled: nil},
	})
	if got != 1 {
		t.Fatalf("got %d", got)
	}
}
