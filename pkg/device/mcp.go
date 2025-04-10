//go:build !tinygo

package device

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

func (ms *MCPServer) doRequest(ctx context.Context, method, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(ms.user, ms.passwd)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("request failed: status %d", resp.StatusCode)
	}

	return body, nil
}

func (ms *MCPServer) handlerGetDevices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body, err := ms.doRequest(ctx, "GET", ms.url+"/devices")
	if err != nil {
		if body != nil {
			return nil, fmt.Errorf("failed to fetch devices: %w: %s", err, string(body))
		}
		return nil, fmt.Errorf("failed to fetch devices: %w", err)
	}

	return mcp.NewToolResultText(string(body)), nil
}

func (ms *MCPServer) toolGetDevices() {
	tool := mcp.NewTool("get_devices",
		mcp.WithDescription("Get all devices running on the Merliot Hub"),
	)
	ms.AddTool(tool, ms.handlerGetDevices)
}

func (ms *MCPServer) handlerAddDevice(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	parentId, _ := request.Params.Arguments["parent_id"].(string)
	if parentId == "" {
		return nil, errors.New("parent-id parameter is required")
	}
	id, _ := request.Params.Arguments["id"].(string)
	if id == "" {
		id = generateRandomId()
	}
	model, _ := request.Params.Arguments["model"].(string)
	if model == "" {
		return nil, errors.New("model parameter is required")
	}
	name, _ := request.Params.Arguments["name"].(string)
	if name == "" {
		return nil, errors.New("name parameter is required")
	}

	// Create URL with query parameters
	reqURL := fmt.Sprintf("%s/create?ParentId=hub&Child.Id=%s&Child.Model=%s&Child.Name=%s",
		ms.url, url.QueryEscape(id), url.QueryEscape(model), url.QueryEscape(name))

	body, err := ms.doRequest(ctx, "POST", reqURL)
	if err != nil {
		if body != nil {
			return nil, fmt.Errorf("failed to create device: %w: %s", err, string(body))
		}
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Device %s created successfully with id %s", name, id)), nil
}

func (ms *MCPServer) toolAddDevice() {
	tool := mcp.NewTool("add_device",
		mcp.WithDescription("Add a new device to the Merliot Hub"),
		mcp.WithString("parent_id",
			mcp.Required(),
			mcp.Description("Parent device ID"),
		),
		mcp.WithString("id",
			mcp.Description("ID of the device (optional, will be generated if not provided)"),
		),
		mcp.WithString("model",
			mcp.Required(),
			mcp.Description("Model of the device"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the device"),
		),
	)
	ms.AddTool(tool, ms.handlerAddDevice)
}

func (ms *MCPServer) handlerRemoveDevice(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := request.Params.Arguments["id"].(string)
	if id == "" {
		return nil, errors.New("id parameter is required")
	}

	reqURL := fmt.Sprintf("%s/destroy?Id=%s", ms.url, url.QueryEscape(id))

	body, err := ms.doRequest(ctx, "DELETE", reqURL)
	if err != nil {
		if body != nil {
			return nil, fmt.Errorf("failed to remove device: %w: %s", err, string(body))
		}
		return nil, fmt.Errorf("failed to remove device: %w", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Device %s removed successfully", id)), nil
}

func (ms *MCPServer) toolRemoveDevice() {
	tool := mcp.NewTool("remove_device",
		mcp.WithDescription("Remove a device from the Merliot Hub"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the device to remove"),
		),
	)
	ms.AddTool(tool, ms.handlerRemoveDevice)
}

func (ms *MCPServer) handlerSave(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body, err := ms.doRequest(ctx, "GET", ms.url+"/save")
	if err != nil {
		if body != nil {
			return nil, fmt.Errorf("failed to save devices: %w: %s", err, string(body))
		}
		return nil, fmt.Errorf("failed to save devices: %w", err)
	}

	return mcp.NewToolResultText("Devices saved successfully"), nil
}

func (ms *MCPServer) toolSave() {
	tool := mcp.NewTool("save",
		mcp.WithDescription("Save current device configuration"),
	)
	ms.AddTool(tool, ms.handlerSave)
}

func (ms *MCPServer) handlerRename(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := request.Params.Arguments["id"].(string)
	if id == "" {
		return nil, errors.New("id parameter is required")
	}
	newName, _ := request.Params.Arguments["new_name"].(string)
	if newName == "" {
		return nil, errors.New("new_name parameter is required")
	}

	reqURL := fmt.Sprintf("%s/rename?Id=%s&NewName=%s",
		ms.url, url.QueryEscape(id), url.QueryEscape(newName))

	body, err := ms.doRequest(ctx, "GET", reqURL)
	if err != nil {
		if body != nil {
			return nil, fmt.Errorf("failed to rename device: %w: %s", err, string(body))
		}
		return nil, fmt.Errorf("failed to rename device: %w", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Device %s renamed to %s successfully", id, newName)), nil
}

func (ms *MCPServer) toolRename() {
	tool := mcp.NewTool("rename",
		mcp.WithDescription("Rename a device on the Merliot Hub"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the device to rename"),
		),
		mcp.WithString("new_name",
			mcp.Required(),
			mcp.Description("New name for the device"),
		),
	)
	ms.AddTool(tool, ms.handlerRename)
}

func (ms *MCPServer) hubResources() {
	// No resources for now
}

func (ms *MCPServer) hubTools() {
	ms.toolGetDevices()
	ms.toolAddDevice()
	ms.toolRemoveDevice()
	ms.toolSave()
	ms.toolRename()
}

func (ms *MCPServer) modelResources(cfg Config) {
}

func (ms *MCPServer) modelTools(cfg Config) {
}

func (ms *MCPServer) build() error {

	ms.hubResources()
	ms.hubTools()

	for _, model := range ms.models {

		// Build device model instance and get config
		device := model.Maker()
		cfg := device.GetConfig()

		// Build MCP resources for model
		ms.modelResources(cfg)

		// Build MCP tools for model
		ms.modelTools(cfg)
	}

	return nil
}
