# Sampatti

MCP server for accessing investment data from Kuvera (Indian mutual funds) and Stockal (US stocks).

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

1. Generate TLS certs (or use a reverse proxy):
   ```bash
   mkdir certs
   # Use certbot, mkcert, or your preferred method
   ```

2. Set `DEV_MODE=false` (or omit it) and configure `BASE_URL` to your public URL.

3. OAuth flow: clients discover auth via `/.well-known/oauth-authorization-server`, register dynamically at `/register`, and authorize at `/authorize` (where users enter platform credentials). Credentials are stored in memory only — never persisted.

4. Connect via the MCP connector API:
   ```json
   {
     "type": "url",
     "url": "https://your-vps.com/mcp",
     "name": "sampatti",
     "authorization_token": "YOUR_OAUTH_TOKEN"
   }
   ```

## Architecture

```
cmd/sampatti/main.go        — entry point
config/config.go            — env configuration
internal/oauth/             — OAuth 2.1 (fosite), middleware, authorize UI
internal/mcp/server.go      — unified MCP server + tool handlers
internal/service/           — Kuvera/Stockal API calls
```

Credentials are never written to disk. Server restart clears all OAuth sessions.
