package cmd

import (
	"errors"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
)

func requireStringFlag(cmd *cobra.Command, name string) error {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return err
	}
	if value == "" {
		return fmt.Errorf("--%s is required", name)
	}
	return nil
}

func meta(accountID string, values map[string]any) map[string]any {
	out := map[string]any{}
	if accountID != "" {
		out["account_id"] = accountID
	}
	for k, v := range values {
		if v != nil && v != "" {
			out[k] = v
		}
	}
	return out
}

func apiError(operation string, err error) error {
	if err == nil {
		return nil
	}
	var reqErr cloudflare.RequestError
	if errors.As(err, &reqErr) {
		ray := reqErr.RayID()
		if ray != "" {
			return fmt.Errorf("%s failed: %w (cf-ray: %s)", operation, err, ray)
		}
	}
	return fmt.Errorf("%s failed: %w", operation, err)
}
