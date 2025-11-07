# Sampatti MCP Server

A Model Context Protocol (MCP) server for accessing investment data from Kuvera and Stockal platforms. Built using the [mcp-go](https://github.com/mark3labs/mcp-go) library and inspired by the [Kite MCP server](https://github.com/zerodha/kite-mcp-server) architecture.

> **Note**: This repository has been refactored from a dashboard application to an MCP server for better integration with AI assistants.

## 🌟 Features

- **Multi-platform support**: Access both Kuvera (mutual funds) and Stockal (US stocks) data
- **Streaming HTTP endpoints**: Separate streaming endpoints for each platform
- **Secure authentication**: Web-based credential input with short-lived tokens
- **Modular architecture**: Separate tools for different platforms
- **MCP Protocol**: Native support for Model Context Protocol

## Architecture

The server runs as a streaming HTTP server with separate MCP endpoints:

- **Kuvera endpoint**: `http://localhost:8081/mcp-kuvera`
- **Stockal endpoint**: `http://localhost:8081/mcp-stockal`
- **Auth UI**: `http://localhost:8080`

Simply start the server and connect your MCP clients to the appropriate endpoints.

## Configuration

Copy `.env.example` to `.env` and configure as needed:

```bash
cp .env.example .env
```

### Environment Variables

- `AUTH_PORT`: Port for authentication web UI (default: 8080)
- `MCP_PORT`: Port for streaming HTTP MCP server (default: 8081)

## Quick Start

1. **Build the server:**
   ```bash
   go build -o sampatti ./cmd/sampatti
   ```

2. **Start the server:**
   ```bash
   ./sampatti
   ```

3. **Generate auth token:**
   - Call `auth_request()` tool to get authentication URL
   - Visit the URL and enter your Kuvera/Stockal credentials
   - Call `get_auth_token()` with the request ID to get your token

4. **Connect MCP clients:**
   - Connect to `http://localhost:8081/mcp-kuvera` for Kuvera tools
   - Connect to `http://localhost:8081/mcp-stockal` for Stockal tools
   - Use the authentication token with the data tools

## Architecture

```
sampatti/
├── cmd/sampatti/           # Main entry point
├── internal/
│   ├── auth/              # Authentication management
│   ├── mcp/               # MCP server implementations
│   │   ├── shared/        # Shared authentication tools
│   │   ├── kuvera/        # Kuvera-specific MCP server
│   │   ├── stockal/       # Stockal-specific MCP server
│   │   └── server.go      # Server manager and routing
│   ├── service/           # Business logic
│   └── web/              # Web authentication UI
└── server/               # Legacy server (deprecated)
```

## Tools Available

### Authentication Tools (Available on both endpoints)
- `auth_request`: Create authentication request and get auth URL
- `get_auth_token`: Retrieve token after user authorization

### Kuvera Tools (Available at `/mcp-kuvera` endpoint)
- `get_kuvera_data`: Fetch portfolio and mutual fund holdings
- `get_gold_silver_rates`: Get gold and silver prices from Kuvera

### Stockal Tools (Available at `/mcp-stockal` endpoint)
- `get_stockal_data`: Fetch US stock portfolio and positions

## Security

- Credentials are entered through a secure web interface
- Short-lived authentication tokens (1 month expiry)
- No credentials stored in environment variables
- HTTP-only for simplicity

## Development

### Prerequisites
- Go 1.25+
- Access to Kuvera and/or Stockal accounts

### Building
```bash
go mod download
go build -o sampatti ./cmd/sampatti
```

### Testing the Server
```bash
# Start the server
./sampatti

# Test endpoints are available at:
# - http://localhost:8081/mcp-kuvera (Kuvera tools)
# - http://localhost:8081/mcp-stockal (Stockal tools)
# - http://localhost:8080 (Auth UI)
```

## Deployment

The server runs as a single streaming HTTP service with separate MCP endpoints:

1. **Single deployment**: One server instance serves both platforms
2. **Separate endpoints**: Connect different clients to different endpoints as needed
3. **Streaming protocol**: Uses HTTP streaming for real-time MCP communication

## License

This project is licensed under the MIT License.