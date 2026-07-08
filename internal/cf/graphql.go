package cf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const graphqlEndpoint = "https://api.cloudflare.com/client/v4/graphql"

type graphqlRequest struct {
	Query string `json:"query"`
}

type graphqlResponse struct {
	Data   json.RawMessage   `json:"data"`
	Errors []json.RawMessage `json:"errors"`
}

// FirewallEventGroup is a grouped firewall event row from GraphQL analytics.
type FirewallEventGroup struct {
	Count      int    `json:"count"`
	Action     string `json:"action"`
	Description string `json:"description"`
	Source     string `json:"source"`
	Host       string `json:"host"`
	RuleID     string `json:"rule_id"`
}

type firewallEventsQuery struct {
	Viewer struct {
		Zones []struct {
			FirewallEventsAdaptiveGroups []struct {
				Count      int `json:"count"`
				Dimensions struct {
					Action                 string `json:"action"`
					Description            string `json:"description"`
					Source                 string `json:"source"`
					ClientRequestHTTPHost  string `json:"clientRequestHTTPHost"`
					RuleID                 string `json:"ruleId"`
				} `json:"dimensions"`
			} `json:"firewallEventsAdaptiveGroups"`
		} `json:"zones"`
	} `json:"viewer"`
}

// ListFirewallEventGroups queries firewallEventsAdaptiveGroups for a zone.
func (c *Client) ListFirewallEventGroups(ctx context.Context, zoneID string, since time.Time, limit int) ([]FirewallEventGroup, error) {
	if limit <= 0 {
		limit = 25
	}
	sinceISO := since.UTC().Format(time.RFC3339)
	query := fmt.Sprintf(`{
  viewer {
    zones(filter: {zoneTag: %q}) {
      firewallEventsAdaptiveGroups(
        limit: %d,
        filter: {datetime_geq: %q},
        orderBy: [count_DESC]
      ) {
        count
        dimensions {
          action
          description
          source
          clientRequestHTTPHost
          ruleId
        }
      }
    }
  }
}`, zoneID, limit, sinceISO)

	body, err := json.Marshal(graphqlRequest{Query: query})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIToken())
	req.Header.Set("Content-Type", "application/json")

	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("graphql request failed: HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}

	var envelope graphqlResponse
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, err
	}
	if len(envelope.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", truncate(string(envelope.Errors[0]), 300))
	}

	var parsed firewallEventsQuery
	if err := json.Unmarshal(envelope.Data, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Viewer.Zones) == 0 {
		return []FirewallEventGroup{}, nil
	}

	rows := parsed.Viewer.Zones[0].FirewallEventsAdaptiveGroups
	out := make([]FirewallEventGroup, 0, len(rows))
	for _, row := range rows {
		out = append(out, FirewallEventGroup{
			Count:       row.Count,
			Action:      row.Dimensions.Action,
			Description: row.Dimensions.Description,
			Source:      row.Dimensions.Source,
			Host:        row.Dimensions.ClientRequestHTTPHost,
			RuleID:      row.Dimensions.RuleID,
		})
	}
	return out, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
