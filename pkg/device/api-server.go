//go:build !tinygo

package device

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *server) setupAPI() {

	// Each device is a ServeMux, routed thru /device/{id}
	s.devices.drange(func(id string, d *device) bool {

		// Install the /device/{id} pattern to point to this device
		s.deviceInstall(d)

		// Add device APIs
		d.installAPI()

		// Install the device packet handlers APIs
		s.packetHandlersInstall(d)

		return true
	})

	// Install / to point to root device
	s.mux.Handle("/", s.root)

	// Install /ws websocket listener, but only if not in demo mode.  In
	// demo mode, we want to ignore any devices dialing in.
	if !s.isSet(flagRunningDemo) {
		s.mux.HandleFunc("/ws", s.wsHandle)
	}

	// Install /wsx websocket listener (wsx is for htmx ws)
	s.mux.HandleFunc("/wsx", s.wsxHandle)

	// Install /wsmcp websocket listener for MCP servers
	s.mux.HandleFunc("/wsmcp", s.wsMcpHandle)

	if s.isSet(flagRunningSite) {
		s.mux.HandleFunc("GET /{$}", s.showSiteHome)
		s.mux.HandleFunc("GET /home", s.showSiteHome)
		s.mux.HandleFunc("GET /home/{page}", s.showSiteHome)
		s.mux.HandleFunc("GET /demo", s.showSiteDemo)
		s.mux.HandleFunc("GET /demo/{page}", s.showSiteDemo)
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
	s.mux.HandleFunc("POST /save", s.saveDevices)
	s.mux.HandleFunc("GET /save-modal", s.showSaveModal)
	s.mux.HandleFunc("POST /create", s.createChild)
	s.mux.HandleFunc("DELETE /destroy", s.destroyChild)
	s.mux.HandleFunc("GET /download-image/{id}", s.downloadImage)
	s.mux.HandleFunc("GET /download-image/{id}/{sessionId}", s.downloadImage)
	s.mux.HandleFunc("GET /deploy-koyeb/{id}/{sessionId}", s.deployKoyeb)
	s.mux.HandleFunc("PUT /rename", s.rename)
	s.mux.HandleFunc("GET /new-modal/{id}", s.showNewModal)
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

// deviceInstall /device/{id} pattern for device
func (s *server) deviceInstall(d *device) {
	prefix := "/device/" + d.Id
	handler := d.deviceHandler(http.StripPrefix(prefix, d))
	s.mux.Handle(prefix+"/", handler)
	s.logInfo("Device installed", "prefix", prefix, "device", d)
}

// modelInstall installs /model/{model} pattern for device model
func (s *server) modelInstall(d *device) {
	prefix := "/model/" + d.Model
	handler := http.StripPrefix(prefix, d)
	s.mux.Handle(prefix+"/", handler)
	s.logInfo("Model installed", "prefix", prefix)
}

func (s *server) installModels() {
	s.models.drange(func(name string, model *Model) bool {
		proto, _ := s.newDevice("proto", name, "proto")
		s.build(proto, s.defaultDeviceFlags())
		proto.installAPI()
		s.modelInstall(proto)
		model.Config = proto.GetConfig()
		return true
	})
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
	devices := s.devices.getPrettyJSON()
	w.Write(devices)
}

func (s *server) saveDevices(w http.ResponseWriter, r *http.Request) {
	if err := s.devicesSave(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *server) showSaveModal(w http.ResponseWriter, r *http.Request) {
	if err := s.root.renderTmpl(w, "modal-save.tmpl", nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *server) newPacketFromRequest(r *http.Request, v any) (*Packet, error) {
	var pkt = s.newPacket().SetPath(r.URL.Path).SetSession(r.Header.Get("session-id"))
	if _, ok := v.(*NoMsg); ok {
		return pkt, nil
	}
	r.ParseForm()
	err := decode(v, r.Form)
	if err != nil {
		return nil, err
	}
	pkt.Msg, err = json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return pkt, nil
}

func (s *server) save() error {
	if s.isSet(flagAutoSave) {
		return s.devicesSave()
	}
	return s.dirty()
}

func validateIds(id, name string) error {
	if err := validateId(id); err != nil {
		return err
	}
	if err := validateName(name); err != nil {
		return err
	}
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

	s.Lock()
	defer s.Unlock()

	parent, ok := s.devices.get(msg.ParentId)
	if !ok {
		return deviceNotFound(msg.ParentId)
	}

	if parent.isSet(flagLocked) && parent == s.root {
		return fmt.Errorf("Create device aborted; parent is locked")
	}

	if err := validateIds(msg.Child.Id, msg.Child.Name); err != nil {
		return err
	}

	return s.addChild(parent, msg.Child.Id, msg.Child.Model, msg.Child.Name, flags)
}

func (s *server) handleCreated(pkt *Packet) {
	if err := s.handleCreate(pkt, flagLocked); err != nil {
		s.logError("Create", "err", err)
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

	if err := s.handleCreate(pkt, 0); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.save(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Route created msg up
	pkt.SetDst(msg.ParentId).SetPath("created").RouteUp()
}

type msgDestroy struct {
	Id string
}

func (s *server) handleDestroy(pkt *Packet) error {
	var msg msgDestroy

	pkt.Unmarshal(&msg)

	s.Lock()
	defer s.Unlock()

	child, ok := s.devices.get(msg.Id)
	if !ok {
		return deviceNotFound(msg.Id)
	}

	if child == s.root {
		return fmt.Errorf("Can't destroy root")
	}

	s.removeChild(child)
	return nil
}

func (s *server) handleDestroyed(pkt *Packet) {
	if err := s.handleDestroy(pkt); err != nil {
		s.logError("Destroy", "err", err)
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

	child, ok := s.devices.get(msg.Id)
	if !ok {
		err := deviceNotFound(msg.Id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if child.isSet(flagGhost) {
		err := deviceNotFound(msg.Id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if child.isSet(flagLocked) {
		http.Error(w, "Can't destroy device; device is locked",
			http.StatusBadRequest)
		return
	}

	if err := s.handleDestroy(pkt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.save(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Route destroyed msg up
	pkt.SetDst(s.root.Id).SetPath("destroyed").RouteUp()
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
		"models": s.childModels(d).unload(),
		"newid":  generateRandomId(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
