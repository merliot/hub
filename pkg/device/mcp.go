//go:build !tinygo

package device

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// ResourceHandlerFunc is a function that handles MCP resource requests
type ResourceHandlerFunc func(ctx context.Context) (string, error)

// MCPServer represents the MCP server for Merliot Hub
type MCPServer struct {
	*mcpserver.MCPServer
	user   string
	passwd string
	url    string
	models Models
}

// MCPServerOption is a MCPServer option
type MCPServerOption func(*MCPServer)

// WithMCPModels returns a MCP ServerOption that sets the device models
// used to build the MCP resources and tools
func WithMCPModels(models Models) MCPServerOption {
	return func(ms *MCPServer) {
		ms.models = models
	}
}

// WithMCPHubURL returns a MCP ServerOption that sets the URL of the Merliot
// Hub.  The MCP server will send http(s) request to the hub URL.  The MCP
// server will also connect to the hub over a websocket dialed on the
// ws(s)://host:port/wsmcp, where host:port are from hub URL.
func WithMCPHubURL(url string) MCPServerOption {
	return func(ms *MCPServer) {
		ms.url = url
	}
}

// WithMCPUser returns a MCP ServerOption that sets the username for HTTP Basic
// Authentication.
func WithMCPUser(user string) MCPServerOption {
	return func(ms *MCPServer) {
		ms.user = user
	}
}

// WithMCPPasswd returns a MCP ServerOption that sets the password for HTTP
// Basic Authentication.
func WithMCPPasswd(passwd string) MCPServerOption {
	return func(ms *MCPServer) {
		ms.passwd = passwd
	}
}

// NewMCPServer creates a new MCPServer instance
func NewMCPServer(options ...MCPServerOption) *MCPServer {

	ms := &MCPServer{
		MCPServer: mcpserver.NewMCPServer(
			"Merliot Hub MCP Server",
			"1.0.0",
		),
	}

	for _, opt := range options {
		opt(ms)
	}

	return ms
}

// ServeStdio starts the stdio server
func (ms *MCPServer) ServeStdio() error {
	if err := ms.build(); err != nil {
		return err
	}
	return mcpserver.ServeStdio(ms.MCPServer)
}

func (ms *MCPServer) handlerGetDevices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", ms.url+"/devices", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth
	req.SetBasicAuth(ms.user, ms.passwd)

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devices: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch devices: status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return mcp.NewToolResultText(string(body)), nil
}

func (ms *MCPServer) toolGetDevices() {
	tool := mcp.NewTool("get_devices",
		mcp.WithDescription("Get all devices running on the Merliot Hub"),
	)
	ms.AddTool(tool, ms.handlerGetDevices)
}

func handlerHelloWorld(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return nil, errors.New("name must be a string")
	}

	return mcp.NewToolResultText(fmt.Sprintf("Hello, %s!", name)), nil
}

func (ms *MCPServer) toolHelloWorld() {
	tool := mcp.NewTool("hello_world",
		mcp.WithDescription("Say hello to someone"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the person to greet"),
		),
	)
	ms.AddTool(tool, handlerHelloWorld)
}

func (ms *MCPServer) hubResources() {
	// No resources for now
}

func (ms *MCPServer) hubTools() {
	ms.toolHelloWorld()
	ms.toolGetDevices()
}

func (ms *MCPServer) modelResources(d *device) {
}

func (ms *MCPServer) modelTools(d *device) {
}

func (ms *MCPServer) build() error {

	ms.hubResources()
	ms.hubTools()

	for _, model := range ms.models {

		// Build device model instance and get config
		d := &device{}
		d.Devicer = model.Maker()
		d.Config = d.GetConfig()

		// Build MCP resources for model
		ms.modelResources(d)

		// Build MCP tools for model
		ms.modelTools(d)
	}

	return nil
}
