package shared

import (
	"context"
	"fmt"

	"github.com/adjaecent/sampatti/internal/auth"
	"github.com/mark3labs/mcp-go/mcp"
)

// AuthToolHandler provides common authentication tools for MCP servers
type AuthToolHandler struct {
	authRequestMgr *auth.AuthRequestManager
	authPort       string
}

func NewAuthToolHandler(authRequestMgr *auth.AuthRequestManager, authPort string) *AuthToolHandler {
	return &AuthToolHandler{
		authRequestMgr: authRequestMgr,
		authPort:       authPort,
	}
}

// GetAuthRequestTool returns the auth_request tool definition
func (h *AuthToolHandler) GetAuthRequestTool() mcp.Tool {
	return mcp.NewTool("auth_request",
		mcp.WithDescription("STEP 1: Create an authentication request that returns a URL for the user to authorize their credentials. Must be called before accessing any investment data."),
	)
}

// GetAuthTokenTool returns the get_auth_token tool definition
func (h *AuthToolHandler) GetAuthTokenTool() mcp.Tool {
	return mcp.NewTool("get_auth_token",
		mcp.WithDescription("STEP 2: Retrieve the authentication token after user has visited the auth URL and entered their credentials"),
		mcp.WithString("request_id",
			mcp.Required(),
			mcp.Description("Authentication request ID from auth_request"),
		),
	)
}

// HandleAuthRequest handles the auth_request tool call
func (h *AuthToolHandler) HandleAuthRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	authRequest, err := h.authRequestMgr.CreateAuthRequest()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create auth request: %v", err)), nil
	}

	authURL := fmt.Sprintf("http://localhost:%s/auth/%s", h.authPort, authRequest.ID)
	
	result := fmt.Sprintf(`🔐 Authentication Request Created

Please visit the following URL to authorize your credentials:
%s

This link expires in 3 minutes for security.
After authorization, use get_auth_token with request_id: %s`, authURL, authRequest.ID)

	return mcp.NewToolResultText(result), nil
}

// HandleGetAuthToken handles the get_auth_token tool call
func (h *AuthToolHandler) HandleGetAuthToken(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	requestID, ok := args["request_id"].(string)
	if !ok || requestID == "" {
		return mcp.NewToolResultError("request_id is required"), nil
	}

	token, expiry, err := h.authRequestMgr.GetAuthToken(requestID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get auth token: %v", err)), nil
	}

	result := fmt.Sprintf(`✅ Authentication Token Retrieved

Token: %s
Expires: %s

You can now use this token with data tools.`, 
		token, expiry.Format("Jan 2, 2006 at 3:04 PM"))

	return mcp.NewToolResultText(result), nil
}

// ValidateToken validates an authentication token and returns credentials
func (h *AuthToolHandler) ValidateToken(token string) (*auth.Credentials, error) {
	if token == "" {
		return nil, fmt.Errorf("authentication token is required. Please first call auth_request() to get an auth URL, have the user enter credentials, then call get_auth_token() to retrieve the token")
	}

	credentials, err := h.authRequestMgr.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %v", err)
	}

	return credentials, nil
}