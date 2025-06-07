# Gomponents Conversion TODO List

This file contains the task breakdown for converting from html/template to gomponents.

## Phase 1: Infrastructure & Simple Components (Low Priority)

- [x] **Static Components**: Convert static components (device-footer.tmpl, robots.txt, static icons/images)
    - device-footer.tmpl was a placeholder (no conversion needed)
    - site-footer.tmpl converted to Go (SiteFooter)
    - robots.txt and images remain static assets
- [ ] **Button Components**: Convert simple button components (button-save.tmpl, button-info.tmpl, button-settings.tmpl)  
- [ ] **Modal Components**: Convert modal components (modal-save.tmpl, modal-new.tmpl, modal-mcp.tmpl)

## Phase 2: Standalone Templates (Medium Priority)

- [ ] **Documentation Pages**: Convert documentation pages (site-docs.tmpl, site-blog.tmpl, /docs/ content)
- [ ] **Device State Templates**: Convert device state templates (online-overview.tmpl, offline-overview.tmpl, created-detail.tmpl, destroyed-detail.tmpl)
- [ ] **Instruction Templates**: Convert instruction templates (instructions-*-parts.tmpl, instructions-*-step1.tmpl)

## Phase 3: Core Device Views (High Priority)

- [ ] **Device Body Templates**: Convert device-specific body templates (body-overview.tmpl, body-detail.tmpl per device type)
- [ ] **Master Page Templates**: Convert master page templates (device.tmpl, session.tmpl with complex composition)

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