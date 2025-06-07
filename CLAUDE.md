# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

```bash
# Run hub from source
go run ./cmd

# Generate models and build required binaries
go generate ./...

# Run all integration tests
go test ./test/...

# Run tests with coverage
go test -v -coverpkg=./pkg/...,./devices/... -coverprofile=coverage.txt ./test/...

# View test coverage
go tool cover -func=coverage.txt

# Docker deployment
docker run -p 8000:8000 merliot/hub

# Build UF2 firmware for specific devices
go run ./cmd/uf2

# Run MCP server for AI integration
go run ./cmd/mcp
```

## Architecture Overview

**Merliot Hub** is a private, AI-integrated IoT device management platform that serves as a gateway between AI systems and physical hardware.

### Core Components

- **Device System** (`pkg/device/`): Core abstractions for device lifecycle, communication, and management
- **Hardware Abstraction** (`pkg/io/`): Platform-specific implementations for GPIO, sensors, actuators
- **Device Implementations** (`devices/`): Concrete device types (sensors, controls, displays)
- **Web Interface**: HTMX-based real-time UI with WebSocket communication
- **MCP Server** (`cmd/mcp/`): Model Context Protocol server for AI integration

### Key Patterns

**Device Interface**: All devices implement `Devicer` with `Setup()`, `Poll()`, `GetConfig()` methods.

**Multi-Platform Support**: Uses Go build tags to support:
- Linux x86-64 (full hub functionality)
- Raspberry Pi (GPIO hardware access)  
- Arduino/microcontrollers (TinyGo firmware)
- Cloud deployment (Koyeb, Docker)

**Packet-Based Communication**: Devices communicate via packet routing system that flows up/down the device hierarchy.

**Factory Pattern**: Device models registered via `Maker` functions in generated `models.go`.

### Development Workflow

1. **Model Generation**: `go generate ./...` rebuilds device registry from `models.json`
2. **Testing**: Integration tests in `test/` cover multi-hub networking and device lifecycle
3. **Firmware Building**: UF2 files generated for drag-and-drop microcontroller deployment
4. **AI Integration**: MCP server enables natural language device control via Claude/Cursor

### File Structure Notes

- `cmd/`: Multiple entry points (hub server, MCP server, firmware builder)
- `pkg/device/`: Core device framework and web serving infrastructure
- `devices/*/`: Individual device implementations with templates and build configs
- `test/`: Integration tests covering full system scenarios
- Templates embedded via `//go:embed` for single-binary deployment

### Platform-Specific Code

Code uses build tags for target-specific implementations:
- `//go:build !tinygo` for full Go environments
- `//go:build tinygo` for microcontroller firmware
- Separate files for platform-specific hardware access (GPIO, sensors)