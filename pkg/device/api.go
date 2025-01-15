//go:build !tinygo

package device

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type APIs map[string]http.HandlerFunc

func (d *device) installAPIs() {

	// Base APIs for all devices

	if runningSite && d == root {
		d.HandleFunc("GET /{$}", d.showSiteHome)
		d.HandleFunc("GET /home", d.showSiteHome)
		d.HandleFunc("GET /home/{page}", d.showSiteHome)
		d.HandleFunc("GET /demo", d.showSiteDemo)
		d.HandleFunc("GET /demo/{page}", d.showSiteDemo)
		d.HandleFunc("GET /status", d.showSiteStatus)
		d.HandleFunc("GET /status/{page}", d.showSiteStatus)
		d.HandleFunc("GET /status/{page}/refresh", d.showSiteStatus)
		d.HandleFunc("GET /doc", d.showSiteDocs)
		d.HandleFunc("GET /doc/{page}", d.showSiteDocs)
		d.HandleFunc("GET /doc/model/{model}", d.showSiteModelDocs)
		d.HandleFunc("GET /blog", d.showSiteBlog)
		d.HandleFunc("GET /blog/{page}", d.showSiteBlog)
	} else {
		d.HandleFunc("GET /{$}", d.showHome)
		d.HandleFunc("GET /home", d.showHome)
		d.HandleFunc("GET /status", d.showStatus)
		d.HandleFunc("GET /status/{page}", d.showStatus)
		d.HandleFunc("GET /status/{page}/refresh", d.showStatus)
		d.HandleFunc("GET /doc", d.showDocs)
		d.HandleFunc("GET /doc/{page}", d.showDocs)
		d.HandleFunc("GET /doc/model/{model}", d.showModelDocs)
	}

	d.HandleFunc("GET /", d.serveStaticFile)

	d.HandleFunc("PUT /nop", d.nop)
	d.HandleFunc("GET /show-view", d.showView)

	d.HandleFunc("GET /state", d.showState)
	d.HandleFunc("GET /code", d.showCode)

	d.HandleFunc("GET /save", d.saveDevices)
	d.HandleFunc("GET /save-modal", d.showSaveModal)

	d.HandleFunc("GET /download-target/{sessionId}", d.showDownloadTarget)
	d.HandleFunc("GET /download-image", d.downloadImage)
	d.HandleFunc("GET /download-image/{sessionId}", d.downloadImage)

	d.HandleFunc("GET /deploy-koyeb/{sessionId}", d.deployKoyeb)

	d.HandleFunc("GET /instructions", d.showInstructions)
	d.HandleFunc("GET /instructions-target", d.showInstructionsTarget)

	d.HandleFunc("GET /edit-name", d.editName)
	d.HandleFunc("GET /rename", d.rename)

	d.HandleFunc("GET /model", d.showModel)

	d.HandleFunc("POST /create", d.createChild)
	d.HandleFunc("DELETE /destroy", d.destroyChild)

	d.HandleFunc("GET /new-modal", d.showNewModal)

	// Device-specific APIs, if any

	if d.APIs != nil {
		for path, fn := range d.APIs {
			d.HandleFunc(path, fn)
		}
	}
}

// modelInstall installs /model/{model} pattern for device in default ServeMux
func (d *device) modelInstall() {
	prefix := "/model/" + d.Model
	handler := basicAuthHandler(http.StripPrefix(prefix, d))
	http.Handle(prefix+"/", handler)
	LogInfo("Model installed", "prefix", prefix)
}

func modelsInstall() {
	for name := range Models {
		model := Models[name]
		proto := &device{Model: name}
		proto._build(model.Maker)
		proto.setupAPI()
		proto.modelInstall()
		model.Config = proto.GetConfig()
		Models[name] = model
	}
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

// deviceInstall installs /device/{id} pattern for device in default ServeMux
func (d *device) deviceInstall() {
	prefix := "/device/" + d.Id
	handler := d.deviceHandler(basicAuthHandler(http.StripPrefix(prefix, d)))
	http.Handle(prefix+"/", handler)
	LogInfo("Device installed", "prefix", prefix, "device", d)
}

func devicesInstall() {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	for _, d := range devices {
		d.deviceInstall()
	}
}

func (d *device) _renderTmpl(w io.Writer, template string, data any) error {
	tmpl := d.templates.Lookup(template)
	if tmpl == nil {
		return fmt.Errorf("Template '%s' not found", template)
	}
	err := tmpl.Execute(w, data)
	if err != nil {
		LogError("Rendering template", "err", err)
	}
	return err
}

func (d *device) renderTmpl(w io.Writer, template string, data any) error {
	d.RLock()
	defer d.RUnlock()
	return d._renderTmpl(w, template, data)
}

func (d *device) _renderSession(w io.Writer, template, sessionId string,
	level int, data map[string]any) error {
	data["sessionId"] = sessionId
	data["level"] = level
	return d._renderTmpl(w, template, data)
}

func (d *device) _render(w io.Writer, sessionId, path, view string,
	level int, data map[string]any) error {

	path = strings.TrimPrefix(path, "/")
	template := path + "-" + view + ".tmpl"

	//LogDebug("_render", "id", d.Id, "session-id", sessionId,
	//	"path", path, "level", level, "template", template)
	if err := d._renderSession(w, template, sessionId, level, data); err != nil {
		return err
	}

	d.saveView(sessionId, view, level)

	return nil
}

func (d *device) _renderWalk(w io.Writer, sessionId, path, view string,
	level int, data map[string]any) error {

	// We're going to walk children, so hold devices lock
	devicesMu.RLock()
	defer devicesMu.RUnlock()

	return d._render(w, sessionId, path, view, level, data)
}

func (d *device) render(w io.Writer, sessionId, path, view string,
	level int, data map[string]any) error {

	d.RLock()
	defer d.RUnlock()

	return d._renderWalk(w, sessionId, path, view, level, data)
}

func (d *device) renderPkt(w io.Writer, sessionId string, pkt *Packet) error {
	var data map[string]any

	view, level := d.lastView(sessionId)
	json.Unmarshal(pkt.Msg, &data)

	if data == nil {
		data = make(map[string]any)
	}

	d.RLock()
	defer d.RUnlock()

	//LogDebug("renderPkt", "id", d.Id, "view", view, "level", level, "pkt", pkt)
	return d._render(w, sessionId, pkt.Path, view, level, data)
}

func (d *device) _renderTemplate(name string, data any) (template.HTML, error) {
	var buf bytes.Buffer
	if err := d._renderTmpl(&buf, name, data); err != nil {
		return template.HTML(""), err
	}
	return template.HTML(buf.String()), nil
}

func RenderTemplate(w io.Writer, id, name string, data any) error {
	d, err := getDevice(id)
	if err != nil {
		return err
	}
	return d.renderTmpl(w, name, data)
}

func (d *device) _renderView(sessionId, path, view string, level int) (template.HTML, error) {
	var buf bytes.Buffer

	if err := d._renderWalk(&buf, sessionId, path, view, level,
		map[string]any{}); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

func (d *device) _renderChildrenWrite(w io.Writer, sessionId string, level int) error {

	if len(d.Children) == 0 {
		return nil
	}

	// Collect child devices from d.Children
	var children []*device
	for _, childId := range d.Children {
		if child, exists := devices[childId]; exists {
			children = append(children, child)
		}
	}

	// TODO allow other sort methods?

	// Sort the collected child devices by ToLower(device.Name)
	sort.Slice(children, func(i, j int) bool {
		return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)
	})

	// Render the child devices in sorted order
	for _, child := range children {
		view, _ := child.lastView(sessionId)
		child.RLock()
		if err := child._render(w, sessionId, "/device", view, level,
			map[string]any{}); err != nil {
			child.RUnlock()
			return err
		}
		child.RUnlock()
	}

	return nil
}

func (d *device) _renderChildren(sessionId string, level int) (template.HTML, error) {
	var buf bytes.Buffer
	err := d._renderChildrenWrite(&buf, sessionId, level)
	return template.HTML(buf.String()), err
}

func (d *device) serveStaticFile(w http.ResponseWriter, r *http.Request) {
	fileExtension := filepath.Ext(r.URL.Path)
	switch fileExtension {
	case ".go", ".tmpl", ".css":
		w.Header().Set("Content-Type", "text/plain")
	case ".gz":
		w.Header().Set("Content-Encoding", "gzip")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	}
	http.FileServer(http.FS(d.layeredFS)).ServeHTTP(w, r)
}

func (d *device) nop(w http.ResponseWriter, r *http.Request) {}

func (d *device) noSessions(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "no more sessions", http.StatusTooManyRequests)
}

func (d *device) showPage(w http.ResponseWriter, r *http.Request,
	template, defaultPage string, pages []page, data map[string]any) {

	data["pages"] = pages
	data["page"] = r.PathValue("page")
	if data["page"] == "" {
		data["page"] = defaultPage
	}

	if err := d.renderTmpl(w, template, data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showSection(w http.ResponseWriter, r *http.Request,
	template, section, defaultPage string, pages []page, data map[string]any) {
	data["section"] = section
	d.showPage(w, r, template, defaultPage, pages, data)
}

func (d *device) showHome(w http.ResponseWriter, r *http.Request) {
	sessionId, ok := newSession()
	if !ok {
		d.noSessions(w, r)
		return
	}
	d.showSection(w, r, "device.tmpl", "home", "", nil, map[string]any{
		"sessionId":  sessionId,
		"pingPeriod": pingPeriod,
	})
}

func (d *device) showStatusRefresh(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	template := "device-status-" + page + ".tmpl"
	if err := d.renderTmpl(w, template, map[string]any{
		"sessions": sessionsStatus(),
		"devices":  devicesStatus(),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showStatus(w http.ResponseWriter, r *http.Request) {
	refresh := path.Base(r.URL.Path)
	if refresh == "refresh" {
		d.showStatusRefresh(w, r)
		return
	}
	d.showSection(w, r, "device.tmpl", "status", "sessions", statusPages, map[string]any{
		"sessions": sessionsStatus(),
		"devices":  devicesStatus(),
	})
}

func (d *device) showDocs(w http.ResponseWriter, r *http.Request) {
	d.showSection(w, r, "device.tmpl", "docs", "quick-start", docPages, map[string]any{
		"models": Models,
		"model":  "",
	})
}

func (d *device) showModelDocs(w http.ResponseWriter, r *http.Request) {
	model := r.PathValue("model")
	d.showSection(w, r, "device.tmpl", "docs", "", docPages, map[string]any{
		"models": Models,
		"model":  model,
	})
}

func (d *device) showView(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	sessionId := r.Header.Get("session-id")
	_, level := d.lastView(sessionId)
	if err := d.render(w, sessionId, "/device", view, level, map[string]any{}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showState(w http.ResponseWriter, r *http.Request) {
	if err := d.renderTmpl(w, "device-state-state.tmpl", nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showCode(w http.ResponseWriter, r *http.Request) {
	// Retrieve top-level entries
	entries, _ := fs.ReadDir(d.layeredFS, ".")
	// Collect entry names
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	d.renderTmpl(w, "code.tmpl", names)
}

func (d *device) saveDevices(w http.ResponseWriter, r *http.Request) {
	if d != root {
		http.Error(w, fmt.Sprintf("Only root device can save"), http.StatusBadRequest)
		return
	}
	if err := devicesSave(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	deviceClean(d.Id)
}

func (d *device) showSaveModal(w http.ResponseWriter, r *http.Request) {
	if err := d.renderTmpl(w, "modal-save.tmpl", nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func showDevices(w http.ResponseWriter, r *http.Request) {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=devices.json")
	state, _ := json.MarshalIndent(devices, "", "\t")
	w.Write(state)
}

func (d *device) deployValues() url.Values {
	values, err := url.ParseQuery(string(d.DeployParams))
	if err != nil {
		panic(err.Error())
	}
	return values
}

func (d *device) selectedTarget(params url.Values) string {
	target := params.Get("target")
	if target == "" {
		target = d.deployValues().Get("target")
	}
	return target
}

func (d *device) showDownloadTarget(w http.ResponseWriter, r *http.Request) {
	selectedTarget := d.selectedTarget(r.URL.Query())
	sessionId := r.PathValue("sessionId")
	err := d.renderTmpl(w, "device-download-target.tmpl", map[string]any{
		"sessionId":      sessionId,
		"selectedTarget": selectedTarget,
		"wantsWifi":      wantsWifi(selectedTarget),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showInstructions(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	template := "instructions-" + view + ".tmpl"
	if err := d.renderTmpl(w, template, nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showInstructionsTarget(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	template := "instructions-" + target + ".tmpl"
	if err := d.renderTmpl(w, template, nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) editName(w http.ResponseWriter, r *http.Request) {
	if err := d.renderTmpl(w, "edit-name.tmpl", d.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

type msgRename struct {
	NewName string
}

func (d *device) rename(w http.ResponseWriter, r *http.Request) {
	var msg msgRename

	pkt, err := newPacketFromRequest(r, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if msg.NewName != "" {
		d.Lock()
		d.Name = msg.NewName
		d.Unlock()
		deviceDirty(root.Id)
		downlinkClose(d.Id)
	}

	// Broadcast /rename msg up
	pkt.SetDst(d.Id).BroadcastUp()
}

func (d *device) apiRouteDown(w http.ResponseWriter, r *http.Request) {
	var msg any

	pkt, err := newPacketFromRequest(r, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pkt.SetDst(d.Id).RouteDown()
}

func (d *device) showModel(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	template := "model-" + view + ".tmpl"
	if err := d.renderTmpl(w, template, d.Config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

type MsgCreated struct {
	Id    string
	Model string
	Name  string
}

func (d *device) createChild(w http.ResponseWriter, r *http.Request) {
	var msg MsgCreated

	if d.isSet(flagLocked) {
		http.Error(w, "Create child aborted; device is locked",
			http.StatusBadRequest)
		return
	}

	pkt, err := newPacketFromRequest(r, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO validate msg.Id, msg.Model, msg.Name

	if err := addChild(d, msg.Id, msg.Model, msg.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Rebuild routing table
	routesBuild(root)

	// Mark root dirty
	deviceDirty(root.Id)

	// Route /created msg down
	pkt.SetDst(d.Id).SetPath("/created").RouteDown()
}

type MsgDestroyed struct {
	ChildId string
}

func (d *device) destroyChild(w http.ResponseWriter, r *http.Request) {
	var msg MsgDestroyed

	pkt, err := newPacketFromRequest(r, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parentId := deviceParent(msg.ChildId)

	if err := removeChild(msg.ChildId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Rebuild routing table
	routesBuild(root)

	// Mark root dirty
	deviceDirty(root.Id)

	// Route /destroyed msg down
	pkt.SetDst(parentId).SetPath("/destroyed").RouteDown()
}

func (d *device) showNewModal(w http.ResponseWriter, r *http.Request) {
	err := d.renderTmpl(w, "modal-new.tmpl", map[string]any{
		"models": d.childModels(),
		"newid":  generateRandomId(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
