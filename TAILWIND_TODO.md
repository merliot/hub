# Tailwind CSS Conversion TODO List

This file contains the task breakdown for converting from custom CSS to Tailwind CSS.

## Phase 1: Infrastructure Setup (High Priority)

- [x] **CDN Integration**: Replace device.css references with Tailwind CDN in template files
  - [x] Update `pkg/device/template/site.tmpl` (line 12)
  - [x] Update `pkg/device/template/device.tmpl` (line 12) 
  - [x] Update `pkg/device/template/sessions.tmpl` (line 10)
  - [x] Test all three template files render correctly

- [x] **Remove CSS File References**: Clean up embedded CSS files and serving logic
  - [x] Remove `css/` directory from `//go:embed` directive in `pkg/device/device-linux.go`
  - [x] Remove CSS file serving logic from `pkg/device/api-device.go`
  - [x] Delete `pkg/device/css/device.css` and `pkg/device/css/device.css.gz`

## Phase 2: Color System Migration (High Priority)

- [x] **Custom Color Audit**: Document all custom color usage
  - [x] List all custom color classes used in templates
  - [x] Map custom colors to Tailwind equivalents:
    - [ ] `.text-violet-creme` (#ddbbff) → `.text-purple-200` or custom
    - [ ] `.bg-sunflower` (#ffcc66) → `.bg-yellow-300` or custom
    - [ ] `.text-african-violet` (#cc88ff) → `.text-purple-400` or custom
    - [ ] `.text-almond-creme` (#ffbbaa) → `.text-orange-200` or custom
    - [ ] `.text-moonlit-violet` (#9944ff) → `.text-purple-600` or custom
    - [ ] `.text-butterscotch` (#ff9966) → `.text-orange-400` or custom
    - [ ] Map remaining 19+ custom colors

- [x] **Color Migration Strategy**: Choose approach for custom colors
  - [x] Option A: Map to closest Tailwind colors and accept visual changes
  - [ ] Option B: Add CSS custom properties for exact color preservation
  - [ ] Option C: Use Tailwind config file with custom color palette

## Phase 3: Template Class Updates (Medium Priority)

- [x] **Layout Class Verification**: Ensure existing classes work with Tailwind
  - [x] Test responsive behavior (grid, flex layouts)
  - [x] Verify spacing utilities work correctly
  - [x] Check sizing utilities (width, height, min/max values)

- [ ] **Device-Specific Template Updates**: Update templates per device type
  - [x] Buttons device templates (`devices/buttons/template/`)
  - [x] Camera device templates (`devices/camera/template/`)
  - [x] GPS device templates (`devices/gps/template/`)
  - [x] Locker device templates (`devices/locker/template/`)
  - [x] ProStar device templates (`devices/prostar/template/`)
  - [x] QR Code device templates (`devices/qrcode/template/`)
  - [x] Relays device templates (`devices/relays/template/`)
  - [x] Temperature device templates (`devices/temp/template/`)
  - [x] Timer device templates (`devices/timer/template/`)

## Phase 4: Component Migration (Medium Priority)

- [x] **Icon Component Migration**: Convert `.icon` styling (main templates)
- [x] **Button Component Migration**: Convert custom button styling (main templates)
- [x] **Modal Component Migration**: Convert modal dialog styling
  - [x] Migrate `modal-new.tmpl` to Tailwind
  - [x] Migrate `modal-save.tmpl` to Tailwind
  - [x] Migrate `modal-mcp.tmpl` to Tailwind
  - [ ] Test modal functionality with new styles

## Phase 5: Advanced Features (Lower Priority)

- [ ] **State-Dependent Styling**: Migrate online/offline state handling
  - [ ] Convert `.offline .panel` rules to conditional classes
  - [ ] Migrate state-based color inversions
  - [ ] Test device state transitions

- [ ] **Documentation Component Migration**: Convert doc-specific styles
  - [ ] Migrate `.cmd-line` styling for command examples
  - [ ] Migrate `.code-snippet` styling for code blocks  
  - [ ] Migrate `.note` styling for documentation notes
  - [ ] Update blog and documentation pages

- [ ] **Custom Utility Migration**: Convert remaining custom utilities
  - [ ] Migrate custom spacing values (`.ml-30`, `.ml-40`, `.ml-50`)
  - [ ] Migrate custom sizing values (`.h-120`, `.h-128`, `.w-120`)
  - [ ] Migrate custom border spacing (`.border-spacing-5-1`)

## Testing & Validation (Ongoing)

- [ ] **Cross-Browser Testing**: Verify styling across browsers
  - [ ] Test in Chrome, Firefox, Safari
  - [ ] Test on mobile devices
  - [ ] Test responsive breakpoints

- [ ] **Device Type Testing**: Test all device interfaces
  - [ ] Test hub device interface
  - [ ] Test all embedded device interfaces
  - [ ] Test demo mode interfaces

- [ ] **Functionality Testing**: Ensure interactive features work
  - [ ] Test HTMX interactions with new styles
  - [ ] Test WebSocket real-time updates
  - [ ] Test modal dialogs and forms
  - [ ] Test responsive navigation and menus

## Notes

- Start with Phase 1 to establish Tailwind CSS foundation
- Document any visual changes that occur during color migration
- Test thoroughly after each phase completion
- Consider creating a style guide for the new Tailwind-based design system
- Keep custom CSS minimal and only for truly unique styling needs