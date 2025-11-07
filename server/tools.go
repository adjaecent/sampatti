package server

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

// registerTools registers all available tools with the MCP server
func (s *SampattiServer) registerTools() {
	// 1. Authentication request tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "auth_request",
		Description: "STEP 1: Create an authentication request that returns a URL for the user to authorize their Kuvera/Stockal credentials. Must be called before accessing any investment data.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}, s.handleAuthRequest)

	// 2. Get authentication token tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_auth_token",
		Description: "STEP 2: Retrieve the authentication token after user has visited the auth URL and entered their credentials",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"request_id": map[string]interface{}{
					"type":        "string",
					"description": "Authentication request ID from auth_request",
				},
			},
			"required": []string{"request_id"},
		},
	}, s.handleGetAuthToken)

	// 3. Kuvera data tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_kuvera_data",
		Description: "STEP 3: Get detailed live data from Kuvera including portfolio and mutual fund holdings. Requires authentication token from get_auth_token.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Authentication token from get_auth_token",
				},
			},
			"required": []string{"token"},
		},
	}, s.handleGetKuveraData)

	// 4. Stockal data tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_stockal_data",
		Description: "Get detailed live data from Stockal (a.k.a Global Investing) including portfolio and US stock positions (requires authentication token)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Authentication token from get_auth_token",
				},
			},
			"required": []string{"token"},
		},
	}, s.handleGetStockalData)

	// 5. Gold/Silver rates tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_gold_silver_rates",
		Description: "Get current gold and silver prices from available sources (requires authentication token)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Authentication token from get_auth_token",
				},
			},
			"required": []string{"token"},
		},
	}, s.handleGetGoldSilverRates)
}

// Tool handler implementations
func (s *SampattiServer) handleAuthRequest(ctx context.Context, request mcp.CallToolRequest) (mcp.CallToolResult, error) {
	authRequest, err := s.authRequestMgr.CreateAuthRequest()
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: fmt.Sprintf("failed to create auth request: %v", err)},
			},
			IsError: true,
		}, nil
	}

	// Check if certificates exist to determine protocol
	protocol := "http"
	if _, err := os.Stat("certs/cert.pem"); err == nil {
		if _, err := os.Stat("certs/key.pem"); err == nil {
			protocol = "https"
		}
	}

	authURL := fmt.Sprintf("%s://localhost:%s/auth/%s", protocol, s.authPort, authRequest.ID)

	result := fmt.Sprintf(`🔐 Authentication Request Created

Please visit the following URL to authorize your credentials:
%s

This link expires in 3 minutes for security.
After authorization, use get_auth_token with request_id: %s`, authURL, authRequest.ID)

	return mcp.CallToolResult{
		Content: []mcp.Content{
			{Type: "text", Text: result},
		},
	}, nil
}

func (s *SampattiServer) handleGetAuthToken(ctx context.Context, request mcp.CallToolRequest) (mcp.CallToolResult, error) {
	requestID, ok := request.Params.Arguments["request_id"].(string)
	if !ok || requestID == "" {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: "request_id is required"},
			},
			IsError: true,
		}, nil
	}

	token, expiry, err := s.authRequestMgr.GetAuthToken(requestID)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: fmt.Sprintf("failed to get auth token: %v", err)},
			},
			IsError: true,
		}, nil
	}

	result := fmt.Sprintf(`✅ Authentication Token Retrieved

Token: %s
Expires: %s

You can now use this token with data tools like get_kuvera_data, get_stockal_data, etc.`,
		token, expiry.Format("Jan 2, 2006 at 3:04 PM"))

	return mcp.CallToolResult{
		Content: []mcp.Content{
			{Type: "text", Text: result},
		},
	}, nil
}

func (s *SampattiServer) handleGetKuveraData(ctx context.Context, request mcp.CallToolRequest) (mcp.CallToolResult, error) {
	token, ok := request.Params.Arguments["token"].(string)
	if !ok || token == "" {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: "authentication token is required. Please first call auth_request() to get an auth URL, have the user enter credentials, then call get_auth_token() to retrieve the token"},
			},
			IsError: true,
		}, nil
	}

	credentials, err := s.authRequestMgr.ValidateToken(token)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: fmt.Sprintf("invalid or expired token: %v", err)},
			},
			IsError: true,
		}, nil
	}

	if credentials.Kuvera == nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: "Kuvera credentials not provided for this token"},
			},
			IsError: true,
		}, nil
	}

	data, err := s.investmentService.FetchKuveraData(credentials.Kuvera.Username, credentials.Kuvera.Password)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: fmt.Sprintf("failed to fetch Kuvera data: %v", err)},
			},
			IsError: true,
		}, nil
	}

	result := fmt.Sprintf("Kuvera Data for %s:\n", data["username"])
	result += "- Portfolio data fetched ✅\n"
	result += "- Holdings data fetched ✅\n"

	if goldPrice := data["gold_price"]; goldPrice != nil {
		result += "- Gold price data available ✅\n"
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{
			{Type: "text", Text: result},
		},
	}, nil
}

func (s *SampattiServer) handleGetStockalData(ctx context.Context, request mcp.CallToolRequest) (mcp.CallToolResult, error) {
	token, ok := request.Params.Arguments["token"].(string)
	if !ok || token == "" {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: "token is required"},
			},
			IsError: true,
		}, nil
	}

	credentials, err := s.authRequestMgr.ValidateToken(token)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: fmt.Sprintf("invalid or expired token: %v", err)},
			},
			IsError: true,
		}, nil
	}

	if credentials.Stockal == nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: "Stockal credentials not provided for this token"},
			},
			IsError: true,
		}, nil
	}

	data, err := s.investmentService.FetchStockalData(credentials.Stockal.Username, credentials.Stockal.Password)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: fmt.Sprintf("failed to fetch Stockal data: %v", err)},
			},
			IsError: true,
		}, nil
	}

	result := fmt.Sprintf("Stockal Data for %s:\n", data["username"])
	result += "- Account summary fetched ✅\n"
	result += "- Portfolio detail fetched ✅\n"

	return mcp.CallToolResult{
		Content: []mcp.Content{
			{Type: "text", Text: result},
		},
	}, nil
}

func (s *SampattiServer) handleGetGoldSilverRates(ctx context.Context, request mcp.CallToolRequest) (mcp.CallToolResult, error) {
	token, ok := request.Params.Arguments["token"].(string)
	if !ok || token == "" {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: "token is required"},
			},
			IsError: true,
		}, nil
	}

	credentials, err := s.authRequestMgr.ValidateToken(token)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: fmt.Sprintf("invalid or expired token: %v", err)},
			},
			IsError: true,
		}, nil
	}

	// Pass available credentials to the service
	var kuveraUsername, kuveraPassword string
	if credentials.Kuvera != nil {
		kuveraUsername = credentials.Kuvera.Username
		kuveraPassword = credentials.Kuvera.Password
	}

	rates, err := s.investmentService.GetGoldSilverRates(kuveraUsername, kuveraPassword)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: fmt.Sprintf("failed to get precious metal rates: %v", err)},
			},
			IsError: true,
		}, nil
	}

	source := rates["source"].(string)
	result := fmt.Sprintf("Precious Metal Rates (Source: %s):\n", source)

	if source == "unavailable" {
		result += "❌ " + rates["error"].(string)
	} else {
		if goldPrice := rates["gold_price"]; goldPrice != nil {
			result += "✅ Gold price data available\n"
		}
		// TODO: Add silver price when available
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{
			{Type: "text", Text: result},
		},
	}, nil
}
