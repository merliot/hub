//go:build !tinygo

package device

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// mcper interface
type mcper interface {
	Desc() string
}

// MCPServer represents the MCP server for Merliot Hub
type MCPServer struct {
	*mcpserver.MCPServer
	user    string
	passwd  string
	url     string
	models  Models
	configs map[string]Config // key: model name
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
		models:  make(Models),
		configs: make(map[string]Config),
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

	if err := ms.mcpWsDial(); err != nil {
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

func (ms *MCPServer) handlerGetModels(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var names []string
	for name := range ms.configs {
		names = append(names, name)
	}
	return mcp.NewToolResultText(strings.Join(names, ",")), nil
}

func (ms *MCPServer) toolGetModels() {
	tool := mcp.NewTool("get_models",
		mcp.WithDescription("Get list of all device models available on the Merliot Hub"),
	)
	ms.AddTool(tool, ms.handlerGetModels)
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
	body, err := ms.doRequest(ctx, "POST", ms.url+"/save")
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

	body, err := ms.doRequest(ctx, "PUT", reqURL)
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

func (ms *MCPServer) handlerGetState(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := request.Params.Arguments["id"].(string)
	if id == "" {
		return nil, errors.New("id parameter is required")
	}
	body, err := ms.doRequest(ctx, "GET", ms.url+"/device/"+id+"/state")
	if err != nil {
		if body != nil {
			return nil, fmt.Errorf("failed to get device state: %w: %s", err, string(body))
		}
		return nil, fmt.Errorf("failed to get device state: %w", err)
	}

	return mcp.NewToolResultText(string(body)), nil
}

func (ms *MCPServer) toolGetState() {
	tool := mcp.NewTool("get_state",
		mcp.WithDescription("Get the state of a device on the Merliot Hub"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the device"),
		),
	)
	ms.AddTool(tool, ms.handlerGetState)
}

func (ms *MCPServer) handlerGetInstructions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := request.Params.Arguments["id"].(string)
	if id == "" {
		return nil, errors.New("id parameter is required")
	}
	target, _ := request.Params.Arguments["target"].(string)
	if target == "" {
		return nil, errors.New("target parameter is required")
	}
	body, err := ms.doRequest(ctx, "GET",
		ms.url+"/device/"+id+"/instructions-target?target="+target)
	if err != nil {
		if body != nil {
			return nil, fmt.Errorf("failed to get device instructions: %w: %s", err, string(body))
		}
		return nil, fmt.Errorf("failed to get device instructions: %w", err)
	}

	return mcp.NewToolResultText(string(body)), nil
}

func (ms *MCPServer) toolGetInstructions() {
	tool := mcp.NewTool("get_instructions",
		mcp.WithDescription("Get the instructions for device on the Merliot Hub.  The instructions include parts list and steps to build, download, and deploy the device."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the device"),
		),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("Build target"),
		),
	)
	ms.AddTool(tool, ms.handlerGetInstructions)
}

func (ms *MCPServer) handlerGetConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	model, _ := request.Params.Arguments["model"].(string)
	if model == "" {
		return nil, errors.New("model parameter is required")
	}

	cfg, ok := ms.configs[model]
	if !ok {
		return nil, errors.New("unknown model '" + model + "'")
	}

	json := cfg.getPrettyJSON()

	return mcp.NewToolResultText(string(json)), nil
}

func (ms *MCPServer) toolGetConfig() {
	tool := mcp.NewTool("get_config",
		mcp.WithDescription("Get the model configuration of a Merliot Hub device model"),
		mcp.WithString("model",
			mcp.Required(),
			mcp.Description("Device model"),
		),
	)
	ms.AddTool(tool, ms.handlerGetConfig)
}

func (ms *MCPServer) handlerGetStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := request.Params.Arguments["id"].(string)
	if id == "" {
		return nil, errors.New("id parameter is required")
	}

	body, err := ms.doRequest(ctx, "GET", ms.url+"/device/"+id+"/status")
	if err != nil {
		if body != nil {
			return nil, fmt.Errorf("failed to get device status: %w: %s", err, string(body))
		}
		return nil, fmt.Errorf("failed to get device status: %w", err)
	}

	return mcp.NewToolResultText(string(body)), nil
}

func (ms *MCPServer) toolGetStatus() {
	tool := mcp.NewTool("get_status",
		mcp.WithDescription("Get the status of a Merliot Hub device.  Device status includes connection status (online/offline)."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Device ID"),
		),
	)
	ms.AddTool(tool, ms.handlerGetStatus)
}

func parseMcpTag(tag string) []mcp.PropertyOption {
	var opts []mcp.PropertyOption

	// Split the tag into components
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch {
		case strings.EqualFold(part, "required"):
			opts = append(opts, mcp.Required())
		case strings.HasPrefix(part, "desc="):
			desc := strings.TrimPrefix(part, "desc=")
			opts = append(opts, mcp.Description(desc))
		}
	}

	return opts
}

func toolOptions(msg any) []mcp.ToolOption {
	var opts []mcp.ToolOption

	m, ok := msg.(mcper)
	if !ok {
		return opts
	}

	opts = append(opts,
		mcp.WithDescription(m.Desc()),
		mcp.WithString("id", mcp.Required(), mcp.Description("Device ID")))

	elem := reflect.ValueOf(msg).Elem()
	for i := 0; i < elem.NumField(); i++ {
		tag := elem.Type().Field(i).Tag.Get("mcp")
		if tag != "" {
			name := strings.ToLower(elem.Type().Field(i).Name)
			popts := parseMcpTag(tag)
			opts = append(opts, mcp.WithString(name, popts...))
		}
	}

	return opts
}

func (ms *MCPServer) handlerCustom(path string, msg any) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		id, _ := request.Params.Arguments["id"].(string)
		if id == "" {
			return nil, errors.New("id parameter is required")
		}

		elem := reflect.ValueOf(msg).Elem()
		pairs := make([]string, elem.NumField())
		for i := 0; i < elem.NumField(); i++ {
			tag := elem.Type().Field(i).Tag.Get("mcp")
			if tag != "" {
				name := elem.Type().Field(i).Name
				lname := strings.ToLower(name)
				val, _ := request.Params.Arguments[lname].(string)
				pairs[i] = name + "=" + val
			}
		}

		params := strings.Join(pairs, "&")

		url := ms.url + "/device/" + id + path + "?" + params
		body, err := ms.doRequest(ctx, "POST", url)
		if err != nil {
			if body != nil {
				return nil, fmt.Errorf("failed to execute custom command: %w: %s", err, string(body))
			}
			return nil, fmt.Errorf("failed to execute custom command: %w", err)
		}

		return mcp.NewToolResultText("Call successful"), nil
	}
}

func (ms *MCPServer) toolCustom(model, path string, msg any) {
	name := model + "_" + path[1:]
	tool := mcp.NewTool(name, toolOptions(msg)...)
	ms.AddTool(tool, ms.handlerCustom(path, msg))
}

func (ms *MCPServer) toolsCustom() {
	for model, cfg := range ms.configs {
		for path, handler := range cfg.PacketHandlers {
			if strings.HasPrefix(path, "/") {
				msg := handler.gen()
				ms.toolCustom(model, path, msg)
			}
		}
	}
}

func (ms *MCPServer) build() error {

	// Cache model configs by making a temp device and saving its config
	for name, model := range ms.models {
		ms.configs[name] = model.Maker().GetConfig()
	}

	ms.toolGetModels()
	ms.toolGetDevices()
	ms.toolAddDevice()
	ms.toolRemoveDevice()
	ms.toolSave()
	ms.toolRename()
	ms.toolGetState()
	ms.toolGetInstructions()
	ms.toolGetConfig()
	ms.toolGetStatus()
	ms.toolsCustom()

	return nil
}
