# Gomponents Conversion TODO List

This file contains the task breakdown for converting from html/template to gomponents.

## Phase 1: Infrastructure & Simple Components (Low Priority)

- [x] **Static Components**: Convert static components (device-footer.tmpl, robots.txt, static icons/images)
    - device-footer.tmpl was a placeholder (no conversion needed)
    - site-footer.tmpl converted to Go (SiteFooter)
    - robots.txt and images remain static assets
- [x] **Button Components**: Convert simple button components (button-save.tmpl, button-info.tmpl, button-settings.tmpl, button-new.tmpl)
    - All button gomponents created (not yet integrated)
- [x] **SiteFooter Integration**: SiteFooter gomponent integrated and tested at /gomponents-site
- [x] **Modal Components**: Create gomponents for modal-save.tmpl, modal-new.tmpl, modal-mcp.tmpl (for later integration)
    - All modal gomponents created and ready for integration
- [In Progress] **Documentation Pages**: Convert documentation pages (site-docs.tmpl, site-blog.tmpl, /docs/ content)

## Phase 2: Standalone Templates (Medium Priority)

- [ ] **Device State Templates**: Convert device state templates (device.tmpl, device-state.tmpl, etc.)

## Phase 3: Full Integration (High Priority)

- [ ] Replace template rendering with gomponents rendering throughout the codebase

## Phase 4: Rendering Infrastructure (High Priority)

- [ ] **Template Function System**: Convert template function system (funcs-linux.go, funcs.go to gomponents helpers)
- [ ] **Dynamic Rendering System**: Convert dynamic rendering system (render.go, layered.go, template inheritance)
- [ ] **Integration Points**: Convert integration points (HTMX integration, WebSocket template updates, session state)

## Notes

- Start with Phase 1 tasks as they have minimal dependencies
- Each phase builds upon the previous one
- Test thoroughly after each phase completion
- Maintain backward compatibility during transition
- Consider using build tags to enable gomponents gradually