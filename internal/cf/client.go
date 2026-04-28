package cf

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

type Options struct {
	APIToken  string
	AccountID string
	Debug     bool
	Timeout   time.Duration
}

type Client struct {
	API       *cloudflare.API
	AccountID string
}

func ResolveOptions(opts Options) Options {
	if opts.APIToken == "" {
		opts.APIToken = firstNonEmpty(os.Getenv("CLOUDFLARE_API_TOKEN"), os.Getenv("CF_API_TOKEN"))
	}
	if opts.AccountID == "" {
		opts.AccountID = firstNonEmpty(os.Getenv("CLOUDFLARE_ACCOUNT_ID"), os.Getenv("CF_ACCOUNT_ID"))
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}
	opts.AccountID = strings.TrimSpace(opts.AccountID)
	return opts
}

func NewClient(opts Options) (*Client, error) {
	opts = ResolveOptions(opts)
	if err := validateAuth(opts); err != nil {
		return nil, err
	}

	httpClient := &http.Client{Timeout: opts.Timeout}
	api, err := cloudflare.NewWithAPIToken(opts.APIToken,
		cloudflare.HTTPClient(httpClient),
		cloudflare.Debug(opts.Debug),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		API:       api,
		AccountID: opts.AccountID,
	}, nil
}

func validateAuth(opts Options) error {
	if strings.TrimSpace(opts.APIToken) == "" {
		return errors.New("missing Cloudflare API token: set CLOUDFLARE_API_TOKEN or CF_API_TOKEN, or pass --api-token")
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// DefaultAccountID returns the configured default account when flags/env set it.
func (c *Client) DefaultAccountID() string {
	return c.AccountID
}

// RequireAccountID returns the default account ID or an error if missing (for account-scoped commands).
func (c *Client) RequireAccountID() (string, error) {
	if c.AccountID == "" {
		return "", errors.New("account ID required: set CLOUDFLARE_ACCOUNT_ID or pass --account-id")
	}
	return c.AccountID, nil
}

// ResolveZoneID returns a zone ID from either a raw ID or by listing zones named domain.
func (c *Client) ResolveZoneID(ctx context.Context, zoneIDOrName string) (string, error) {
	zoneIDOrName = strings.TrimSpace(zoneIDOrName)
	if zoneIDOrName == "" {
		return "", errors.New("zone ID or zone name is required")
	}
	if looksLikeZoneID(zoneIDOrName) {
		return zoneIDOrName, nil
	}
	zones, err := c.API.ListZones(ctx, zoneIDOrName)
	if err != nil {
		return "", err
	}
	if len(zones) == 0 {
		return "", errors.New("no zone found for name " + zoneIDOrName)
	}
	if len(zones) > 1 {
		return "", errors.New("multiple zones matched name; pass --zone-id explicitly")
	}
	return zones[0].ID, nil
}

func looksLikeZoneID(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		if r >= 'a' && r <= 'f' {
			continue
		}
		return false
	}
	return true
}
