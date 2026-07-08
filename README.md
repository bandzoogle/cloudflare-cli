# Cloudflare CLI

`cfcli` is a read-only Cloudflare CLI designed for scripts and LLM agents. Commands stay narrow and discoverable while returning stable JSON on stdout.

## Auth

Use an API token with read-only permissions for the routes you need (see [Required permissions](#required-permissions)):

```sh
export CLOUDFLARE_API_TOKEN=...
# or
export CF_API_TOKEN=...
```

Optional default account for account-scoped commands (for example `workers list`):

```sh
export CLOUDFLARE_ACCOUNT_ID=...
# or
export CF_ACCOUNT_ID=...
```

Flags override environment variables when set:

```sh
cfcli --api-token "$CLOUDFLARE_API_TOKEN" workers list --account-id "$CLOUDFLARE_ACCOUNT_ID"
```

## Output

Every read command writes JSON to stdout. Diagnostics and errors go to stderr.

Default output is wrapped:

```json
{
  "query": {
    "command": "zones list",
    "name": ""
  },
  "meta": {
    "account_id": "...",
    "count": 3
  },
  "data": []
}
```

Use `--pretty` for indented JSON and `--raw` to print vendor-shaped data without the `query` and `meta` wrapper.

## Commands

```sh
cfcli zones list
cfcli zones list --name example.com
cfcli zones get ZONE_ID

cfcli zones settings list --zone-name example.com
cfcli zones settings get email_obfuscation --zone-name example.com

cfcli dns list --zone-id ZONE_ID
cfcli dns list --zone-name example.com --type A
cfcli dns get --zone-id ZONE_ID --record-id RECORD_ID

cfcli accounts list
cfcli accounts list --name prod
cfcli accounts get ACCOUNT_ID

cfcli workers list
cfcli workers list --account-id ACCOUNT_ID

cfcli user whoami

cfcli scopes
cfcli scopes --command dns
```

## Required permissions

Use `cfcli scopes` to print the Cloudflare API token permission groups used by each command. This command does not call the Cloudflare API.

Cloudflare grants access through granular token permissions in the dashboard (**My Profile → API Tokens**). Create a custom token with read-only scopes matching the routes you use.

Unlike most commands, `cfcli scopes` prints a compact human-readable list by default because it is primarily a setup reference. Use `--raw` for the JSON envelope.

## MCP tradeoff

This CLI complements MCP rather than replacing it everywhere. MCP helps with integrated auth and rich tool discovery, but adds server and tool context to conversations. A CLI gives agents a smaller contract: `--help`, explicit commands, JSON stdout, stderr diagnostics, and replayable invocations outside Cursor.

## Binary releases

Pushes to `main` that pass CI and change Go sources or `go.mod` / `go.sum` trigger a patch semver GitHub release (for example `v0.1.1`). Builds are published for Linux and macOS on `amd64` and `arm64`.

Stable download URLs (always point at the **latest** release’s asset names):

- Linux x86_64: `https://github.com/bandzoogle/cloudflare-cli/releases/latest/download/cfcli_linux_amd64.tar.gz`
- Linux arm64: `https://github.com/bandzoogle/cloudflare-cli/releases/latest/download/cfcli_linux_arm64.tar.gz`
- macOS x86_64: `https://github.com/bandzoogle/cloudflare-cli/releases/latest/download/cfcli_darwin_amd64.tar.gz`
- macOS arm64: `https://github.com/bandzoogle/cloudflare-cli/releases/latest/download/cfcli_darwin_arm64.tar.gz`

Each archive contains a single `cfcli` binary. Run `cfcli --version` to see the release tag baked into the binary. Versioned archives (`cfcli_<tag>_linux_amd64.tar.gz`, etc.) are attached for pinning.

## Development

```sh
bin/build
```

`bin/build` runs unit tests, builds `dist/cfcli`, and smoke-tests help output for the main read-only command groups. Live Cloudflare smoke tests require credentials and are intentionally manual.
