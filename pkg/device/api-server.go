//go:build !tinygo

package device

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
)

func (s *server) setupAPI() {

	// Each device is a ServeMux, routed thru /device/{id}
	s.devices.drange(func(id string, d *device) bool {

		// Install the /device/{id} pattern to point to this device
		d.install()

		// Add device APIs
		d.installAPI()

		// Install the device packet handlers APIs
		s.packetHandlersInstall(d)

		return true
	})

	// Install /model/{model} patterns for models
	s.installModels()

	// Install / to point to root device
	s.mux.Handle("/", s.root)

	// Install /ws websocket listener, but only if not in demo mode.  In
	// demo mode, we want to ignore any devices dialing in.
	if !s.runningDemo {
		s.mux.HandleFunc("/ws", s.wsHandle)
	}

	// Install /wsx websocket listener (wsx is for htmx ws)
	s.mux.HandleFunc("/wsx", s.wsxHandle)

	if s.runningSite {
		s.mux.HandleFunc("GET /{$}", s.showSiteHome)
		s.mux.HandleFunc("GET /home", s.showSiteHome)
		s.mux.HandleFunc("GET /home/{page}", s.showSiteHome)
		s.mux.HandleFunc("GET /demo", s.showSiteDemo)
		s.mux.HandleFunc("GET /demo/{page}", s.showSiteDemo)
		s.mux.HandleFunc("GET /status", s.showSiteStatus)
		s.mux.HandleFunc("GET /status/{page}", s.showSiteStatus)
		s.mux.HandleFunc("GET /status/{page}/refresh", s.showSiteStatus)
		s.mux.HandleFunc("GET /doc", s.showSiteDocs)
		s.mux.HandleFunc("GET /doc/{page}", s.showSiteDocs)
		s.mux.HandleFunc("GET /blog", s.showSiteBlog)
		s.mux.HandleFunc("GET /blog/{page}", s.showSiteBlog)

	} else {
		s.mux.HandleFunc("GET /{$}", s.showHome)
		s.mux.HandleFunc("GET /home", s.showHome)
	}

	s.mux.HandleFunc("GET /devices", s.showDevices)
	s.mux.HandleFunc("PUT /nop", func(w http.ResponseWriter, r *http.Request) {})
	s.mux.HandleFunc("GET /save", s.saveDevices)
	s.mux.HandleFunc("GET /save-modal", s.showSaveModal)
	s.mux.HandleFunc("POST /create", s.createChild)
	s.mux.HandleFunc("DELETE /destroy", s.destroyChild)
	s.mux.HandleFunc("GET /download-image/{id}", s.downloadImage)
	s.mux.HandleFunc("GET /download-image/{id}/{sessionId}", s.downloadImage)
	s.mux.HandleFunc("GET /deploy-koyeb/{id}/{sessionId}", s.deployKoyeb)
	s.mux.HandleFunc("GET /rename", s.rename)
	s.mux.HandleFunc("GET /new-modal/{id}", s.showNewModal)
}

// modelInstall installs /model/{model} pattern for device model
func (s *server) modelInstall(d *device) {
	prefix := "/model/" + d.Model
	handler := http.StripPrefix(prefix, d)
	s.mux.Handle(prefix+"/", handler)
	LogInfo("Model installed", "prefix", prefix)
}

func (d *device) deviceHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if d.isSet(flagGhost) {
			// Ignore API request for ghost devices
			http.Error(w, "Device is a Ghost", http.StatusGone)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// install /device/{id} pattern for device
func (d *device) install() {
	prefix := "/device/" + d.Id
	handler := d.deviceHandler(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
	LogInfo("Device installed", "prefix", prefix, "device", d)
}

func (s *server) installModels() {
	for name := range s.models {
		model := s.models[name]
		proto := &device{
			Model: name,
			model: model,
		}
		proto.build(model, s.flags())
		proto.installAPI()
		s.modelInstall(proto)
		model.Config = proto.GetConfig()
		s.models[name] = model
	}
}

func (s *server) showHome(w http.ResponseWriter, r *http.Request) {
	sessionId, ok := s.sessions.newSession()
	if !ok {
		s.sessions.noSessions(w, r)
		return
	}
	w.Header().Set("session-id", sessionId)
	s.root.showSection(w, r, "device.tmpl", "home", "", nil, map[string]any{
		"sessionId":  sessionId,
		"pingPeriod": s.wsxPingPeriod,
	})
}

func (s *server) showDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=devices.json")
	devices, _ := json.MarshalIndent(s.devices.getJSON(), "", "\t")
	w.Write(devices)
}

func (s *server) saveDevices(w http.ResponseWriter, r *http.Request) {
	if err := s.devicesSave(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	s.root.clean()
}

func (s *server) showSaveModal(w http.ResponseWriter, r *http.Request) {
	if err := s.root.renderTmpl(w, "modal-save.tmpl", nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *server) showStatusRefresh(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	template := "device-status-" + page + ".tmpl"
	if err := s.root.renderTmpl(w, template, map[string]any{
		"sessions": s.sessions.status(),
		"devices":  s.devices.status(),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *server) showStatus(w http.ResponseWriter, r *http.Request) {
	refresh := path.Base(r.URL.Path)
	if refresh == "refresh" {
		s.showStatusRefresh(w, r)
		return
	}
	s.root.showSection(w, r, "device.tmpl", "status", "sessions", statusPages, map[string]any{
		"sessions": s.sessions.status(),
		"devices":  s.devices.status(),
	})
}

func (s *server) save() error {
	var autoSave = Getenv("AUTO_SAVE", "true") == "true"

	if autoSave {
		return s.devicesSave()
	}

	// Mark root dirty so user can manually save
	s.root.dirty()

	return nil
}

type msgCreated struct {
	ParentId string
	Child    struct {
		Id    string
		Model string
		Name  string
	}
}

func (s *server) handleCreate(pkt *Packet, flags flags) error {
	var msg msgCreated

	pkt.Unmarshal(&msg)

	parent, ok := s.devices.get(msg.ParentId)
	if !ok {
		return deviceNotFound(msg.ParentId)
	}

	if parent.isSet(flagLocked) {
		return fmt.Errorf("Create device aborted; parent is locked")
	}

	if err := validateIds(msg.Child.Id, msg.Child.Name); err != nil {
		return err
	}

	return s.addChild(parent, msg.Child.Id, msg.Child.Model, msg.Child.Name, flags)
}

func (s *server) handleCreated(pkt *Packet) {
	if err := s.handleCreate(pkt, flagLocked); err != nil {
		LogError("Create", "err", err)
		return
	}
	pkt.BroadcastUp()
}

func (s *server) createChild(w http.ResponseWriter, r *http.Request) {
	var msg msgCreated

	pkt, err := s.newPacketFromRequest(r, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.handleCreate(pkt, 0); err == nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.save(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Route /created msg up
	pkt.SetPath("/created").RouteUp()
}

type msgDestroy struct {
	Id string
}

func (s *server) handleDestroy(pkt *Packet) error {
	var msg msgDestroy

	pkt.Unmarshal(&msg)

	child, ok := s.devices.get(msg.Id)
	if !ok {
		return deviceNotFound(msg.Id)
	}

	if child.isSet(flagLocked) {
		return fmt.Errorf("Can't destroy device; device is locked")
	}

	if child == s.root {
		return fmt.Errorf("Can't destroy root")
	}

	s.removeChild(child)
	return nil
}

func (s *server) handleDestroyed(pkt *Packet) {
	if err := s.handleDestroy(pkt); err != nil {
		LogError("Destroy", "err", err)
		return
	}
	pkt.BroadcastUp()
}

func (s *server) destroyChild(w http.ResponseWriter, r *http.Request) {
	var msg msgDestroy

	pkt, err := s.newPacketFromRequest(r, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.handleDestroy(pkt); err == nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.save(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Route /destroyed msg up
	pkt.SetPath("/destroyed").RouteUp()
}

type msgRename struct {
	Id      string
	NewName string
}

func (s *server) rename(w http.ResponseWriter, r *http.Request) {

	var msg msgRename

	pkt, err := s.newPacketFromRequest(r, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d, ok := s.devices.get(msg.Id)
	if !ok {
		http.Error(w, deviceNotFound(msg.Id).Error(), http.StatusBadRequest)
		return
	}

	if err := validateName(msg.NewName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d.Name = msg.NewName

	if err := s.save(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Kick device downlink offline (if online).  Downlink device will try
	// to connect with the wrong name, which will fail.  Downlink device
	// needs to be updated with new image with corrected name.
	s.downlinks.linkClose(d.Id)

	// Broadcast /rename msg up
	pkt.SetDst(d.Id).BroadcastUp()
}

func (s *server) showNewModal(w http.ResponseWriter, r *http.Request) {

	var id = r.PathValue("id")

	d, exists := s.devices.get(id)
	if !exists {
		err := fmt.Errorf("Can't show new modal dialog: unknown device id '%s'", id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := d.renderTmpl(w, "modal-new.tmpl", map[string]any{
		"models": s.childModels(d),
		"newid":  generateRandomId(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
