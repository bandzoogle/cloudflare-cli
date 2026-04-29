package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bandzoogle/cloudflare-cli/internal/cf"
	"github.com/bandzoogle/cloudflare-cli/internal/output"
	"github.com/spf13/cobra"
)

type globalOptions struct {
	apiToken  string
	accountID string
	timeout   time.Duration
	debug     bool
	pretty    bool
	raw       bool
}

var globals globalOptions

var rootCmd = &cobra.Command{
	Use:     "cfcli",
	Version: Version,
	Short:   "Read-only Cloudflare CLI for scripts and LLM agents",
	Long: `cfcli wraps read-only Cloudflare API calls with stable JSON output.

Authenticate with an API token via CLOUDFLARE_API_TOKEN or CF_API_TOKEN.`,
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&globals.apiToken, "api-token", "", "Cloudflare API token (prefer CLOUDFLARE_API_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&globals.accountID, "account-id", "", "Default account ID for account-scoped commands (prefer CLOUDFLARE_ACCOUNT_ID)")
	rootCmd.PersistentFlags().DurationVar(&globals.timeout, "timeout", 30*time.Second, "HTTP request timeout")
	rootCmd.PersistentFlags().BoolVar(&globals.debug, "debug", false, "Enable Cloudflare client debug logging")
	rootCmd.PersistentFlags().BoolVar(&globals.pretty, "pretty", false, "Pretty-print JSON output")
	rootCmd.PersistentFlags().BoolVar(&globals.raw, "raw", false, "Print vendor response JSON without the cfcli envelope")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cfClient(*cobra.Command) (*cf.Client, error) {
	return cf.NewClient(cf.Options{
		APIToken:  globals.apiToken,
		AccountID: globals.accountID,
		Debug:     globals.debug,
		Timeout:   globals.timeout,
	})
}

func outputOptions() output.Options {
	return output.Options{
		Pretty: globals.pretty,
		Raw:    globals.raw,
	}
}

func commandContext() context.Context {
	return context.Background()
}
