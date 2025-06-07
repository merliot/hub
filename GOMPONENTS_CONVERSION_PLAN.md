# Conversion Plan: html/template → gomponents

## Overview

This document outlines the strategy for converting Merliot Hub's template system from Go's html/template package to github.com/maragudk/gomponents. The conversion is planned in logical phases to minimize disruption while building toward a type-safe, component-based UI system.

## Current Template System Architecture

### Template Definition and Loading

**Core Template Management**:
- Templates are defined in `.tmpl` files organized hierarchically across the codebase
- Base templates are embedded in `pkg/device/template/` (100+ templates)
- Device-specific templates are in `devices/{device}/template/`
- Templates are embedded at compile time using `//go:embed` directives

**Layered File System (`layered.go`)**:
- Uses a sophisticated layered filesystem approach via `layeredFS` struct
- Allows device-specific templates to override base templates
- Template resolution: device-specific FS stacked on top of base `deviceFs`
- When looking for `foo.tmpl`, device-specific version wins over base version

### Template Functions Implementation

**Function Map Architecture**:
- Three-tier function system: server functions → device functions → custom functions
- Base server functions (`baseFuncs()`) in `pkg/device/funcs-linux.go`
- Base device functions (`baseFuncs()`) provide device-specific context
- Custom device functions via `Config.FuncMap` allow specialized template functions

**Key Template Functions**:
```go
// Server-level functions
"saveToClipboard", "devicesJSON", "toLower", "title", "add", "mult", 
"joinStrings", "contains", "tinygoTarget", "ssids", "bodyColors", "isDirty"

// Device-level functions  
"id", "model", "name", "uniq", "state", "stateJSON", "uptime", "targets",
"isRoot", "isOnline", "isLocked", "renderTemplate", "renderView", "renderChildren"
```

### Template Rendering Entry Points

**Main Rendering Functions** (`render.go`):
- `renderTmpl()`: Core template execution with error handling
- `render()`: Path/view-based rendering with template name construction
- `renderPkt()`: Packet-driven rendering for real-time updates
- `renderTemplate()`: Returns `template.HTML` for nested template calls
- `renderView()`: Device view rendering with session/level context
- `renderChildren()`: Recursive child device rendering

**Template Resolution Pattern**:
```go
template := path + "-" + view + ".tmpl"
// Examples: "device-detail.tmpl", "device-overview.tmpl", "clicked-state.tmpl"
```

## Conversion Plan

### **Phase 1: Infrastructure & Simple Components (Easiest)**

1. **Static Components** - No dynamic data, pure HTML
   - `device-footer.tmpl` 
   - `robots.txt` serving
   - Static image/icon components
   - CSS class utilities

2. **Simple Button Components** - Minimal template functions
   - `button-save.tmpl` (uses `uniq`, `isDirty`, `saveToClipboard`)
   - `button-info.tmpl`, `button-settings.tmpl`, etc.
   - Icons and small UI elements

3. **Modal Components** - Self-contained dialogs
   - `modal-save.tmpl`, `modal-new.tmpl`, `modal-mcp.tmpl`
   - These are typically isolated from complex device state

### **Phase 2: Standalone Templates (Medium)**

4. **Documentation Pages** - Static content with minimal functions
   - `site-docs.tmpl`, `site-blog.tmpl` 
   - Most content in `/docs/` and `/blog/` directories
   - Limited to basic template functions like `title`, `contains`

5. **Device State Templates** - Single-purpose state renderers  
   - `online-overview.tmpl`, `offline-overview.tmpl`
   - `created-detail.tmpl`, `destroyed-detail.tmpl`
   - These use device state but are simpler than full device views

6. **Instruction Templates** - Platform-specific setup guides
   - `instructions-*-parts.tmpl`, `instructions-*-step1.tmpl`
   - Device-specific but relatively simple structure

### **Phase 3: Core Device Views (Harder)**

7. **Device-Specific Body Templates** - Custom device content
   - `body-overview.tmpl`, `body-detail.tmpl` per device type
   - Device-specific state rendering (buttons, GPS, camera, etc.)
   - Template functions: `state`, `stateJSON`, `model`, `uniq`

8. **Master Page Templates** - Complex composition
   - `device.tmpl` - Main HTML shell with header/footer composition
   - `session.tmpl` - Session management and WebSocket integration
   - Uses `renderTemplate`, `joinStrings`, `bodyColors`

### **Phase 4: Rendering Infrastructure (Hardest)**

9. **Template Function System** - Core template functions
   - Convert `funcs-linux.go` and `funcs.go` to gomponents helpers
   - Functions like `renderTemplate`, `renderView`, `renderChildren`
   - State management: `isDirty`, `isOnline`, `isLocked`

10. **Dynamic Rendering System** - Core rendering engine
    - `render.go` - Template resolution and execution
    - `layered.go` - Layered filesystem and template inheritance
    - Device hierarchy rendering with `renderChildren`

11. **Integration Points** - WebSocket and real-time updates
    - HTMX integration patterns
    - WebSocket template updates
    - Session state management

## Strategic Considerations

### Compatibility Strategy
- Keep both systems running in parallel during transition
- Use build tags or feature flags to enable gomponents gradually
- Start with leaf components that don't depend on template inheritance

### Template Function Migration
- Many template functions (`uniq`, `state`, `model`) become Go helper functions
- Complex functions like `renderTemplate` need architectural changes
- Template composition becomes Go function composition

### Data Flow Changes
- Template data maps become strongly-typed Go structs
- Template inheritance becomes Go function composition
- Dynamic template selection becomes conditional rendering logic

### Benefits After Conversion
- Type safety and compile-time template checking
- Better IDE support and refactoring capabilities  
- Elimination of template string parsing at runtime
- More natural Go patterns for component composition

## Implementation Notes

This plan prioritizes **low-risk, high-value conversions first** while building toward the more complex architectural changes needed for the core rendering system. Each phase builds upon the previous one, allowing for incremental validation and testing throughout the conversion process.