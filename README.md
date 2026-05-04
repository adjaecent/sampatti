# Sampatti

MCP server for accessing investment data from [Kuvera](https://github.com/adjaecent/unofficial-kuvera-api) (Indian mutual funds) and [Stockal](https://github.com/adjaecent/unofficial-stockal-api) (US stocks).

## Tools

- `get_kuvera_portfolio` — portfolio value, gains, XIRR, asset breakdown
- `get_kuvera_holdings` — individual fund details, SIPs, folio numbers
- `get_gold_price` — current gold buy/sell prices
- `get_stockal_account` — cash balances, trading restrictions
- `get_stockal_holdings` — US stock positions, prices, gain/loss

## Setup

```bash
cp .env.example .env
# Edit .env with your credentials

go build -o sampatti ./cmd/sampatti
./sampatti
```

## Configuration

See `.env.example` for all options:

- `PORT` — server port (default: 8081)
- `BASE_URL` — public URL for OAuth metadata
- `DEV_MODE=true` — skip OAuth, use env credentials directly (for local dev)
- `KUVERA_USERNAME` / `KUVERA_PASSWORD` — Kuvera credentials (dev mode)
- `STOCKAL_USERNAME` / `STOCKAL_PASSWORD` — Stockal credentials (dev mode)
- `OAUTH_SECRET` — hex-encoded HMAC secret (production; random if omitted)

## Local dev (Claude Code)

```bash
# Start server with DEV_MODE=true in .env
go run ./cmd/sampatti

# Add to Claude Code
claude mcp add --transport http sampatti https://localhost:8081/mcp
```

## Production (VPS)

1. Deploy behind a reverse proxy (e.g. Caddy) for TLS termination.

2. Set `DEV_MODE=false` (or omit it) and configure `BASE_URL` to your public URL.

3. OAuth flow: clients discover auth via `/.well-known/oauth-authorization-server`, register dynamically at `/register`, and authorize at `/authorize` (where users enter platform credentials).

4. Add as an MCP connector in Claude (or any MCP client) pointing to `https://your-domain.com/mcp`.

## Credential storage

Credentials and OAuth sessions are never written to disk. They are encrypted (AES-256-GCM, keyed from `OAUTH_SECRET`) and stored on `/dev/shm` — host memory backed by tmpfs. This means they survive process restarts and container recreation, but are wiped on host reboot. The Docker container bind-mounts the host's `/dev/shm` for this purpose.

## Architecture

```
cmd/sampatti/main.go        — entry point
config/config.go            — env configuration
internal/oauth/             — OAuth 2.1 (fosite), middleware, authorize UI
internal/mcp/server.go      — unified MCP server + tool handlers
internal/service/           — Kuvera/Stockal API calls
```
