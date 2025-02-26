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
	"path/filepath"
	"strings"
)

type APIs map[string]http.HandlerFunc

func (d *device) setupAPI() {
	// Install base + device APIs
	d.installAPIs()
	// Install the device-specific packet handlers APIs
	d.packetHandlersInstall()
}

func (d *device) installAPIs() {

	// Base APIs for all devices

	d.HandleFunc("GET /", d.serveStaticFile)

	d.HandleFunc("GET /show-view", d.showView)

	d.HandleFunc("GET /state", d.showState)
	d.HandleFunc("GET /code", d.showCode)

	d.HandleFunc("GET /download-target/{sessionId}", d.showDownloadTarget)
	d.HandleFunc("GET /download-image", d.downloadImage)
	d.HandleFunc("GET /download-image/{sessionId}", d.downloadImage)

	d.HandleFunc("GET /deploy-koyeb/{sessionId}", d.deployKoyeb)

	d.HandleFunc("GET /instructions", d.showInstructions)
	d.HandleFunc("GET /instructions-target", d.showInstructionsTarget)

	d.HandleFunc("GET /edit-name", d.editName)

	d.HandleFunc("GET /model", d.showModel)

	d.HandleFunc("GET /new-modal", d.showNewModal)

	// Device-specific APIs, if any

	if d.APIs != nil {
		for path, fn := range d.APIs {
			d.HandleFunc(path, fn)
		}
	}
}

func (d *device) renderTmpl(w io.Writer, template string, data any) error {
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

func (d *device) renderSession(w io.Writer, template, sessionId string,
	level int, data map[string]any) error {
	data["sessionId"] = sessionId
	data["level"] = level
	return d.renderTmpl(w, template, data)
}

func (d *device) render(w io.Writer, sessionId, path, view string,
	level int, data map[string]any) error {

	path = strings.TrimPrefix(path, "/")
	template := path + "-" + view + ".tmpl"

	//LogDebug("_render", "id", d.Id, "session-id", sessionId,
	//	"path", path, "level", level, "template", template)
	if err := d.renderSession(w, template, sessionId, level, data); err != nil {
		return err
	}

	d.saveView(sessionId, view, level)

	return nil
}

func (d *device) renderPkt(w io.Writer, sessionId string, pkt *Packet) error {
	var data map[string]any

	view, level := d.lastView(sessionId)
	json.Unmarshal(pkt.Msg, &data)

	if data == nil {
		data = make(map[string]any)
	}

	//LogDebug("renderPkt", "id", d.Id, "view", view, "level", level, "pkt", pkt)
	return d.render(w, sessionId, pkt.Path, view, level, data)
}

func (d *device) renderTemplate(name string, data any) (template.HTML, error) {
	var buf bytes.Buffer
	if err := d.renderTmpl(&buf, name, data); err != nil {
		return template.HTML(""), err
	}
	return template.HTML(buf.String()), nil
}

/*
func RenderTemplate(w io.Writer, id, name string, data any) error {
	d, err := getDevice(id)
	if err != nil {
		return err
	}
	return d.renderTmpl(w, name, data)
}
*/

func (d *device) renderView(sessionId, path, view string, level int) (template.HTML, error) {
	var buf bytes.Buffer

	if err := d.render(&buf, sessionId, path, view, level,
		map[string]any{}); err != nil {
		return template.HTML(""), err
	}

	return template.HTML(buf.String()), nil
}

func (d *device) renderChildrenWrite(w io.Writer, sessionId string, level int) error {
	return d.children.sortedByName(func(id string, child *device) error {
		view, _ := child.lastView(sessionId)
		if err := child.render(w, sessionId, "/device", view, level,
			map[string]any{}); err != nil {
			return err
		}
		return nil
	})
}

func (d *device) renderChildren(sessionId string, level int) (template.HTML, error) {
	var buf bytes.Buffer
	err := d.renderChildrenWrite(&buf, sessionId, level)
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

func (s *server) showDocs(w http.ResponseWriter, r *http.Request) {
	s.root.showSection(w, r, "device.tmpl", "docs", "quick-start", docPages, map[string]any{
		"models": s.models,
		"model":  "",
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

func wantsWifi(target string) bool {
	return target == "pyportal" || target == "wioterminal" || target == "nano-rp2040"
}

func wantsHttpPort(target string) bool {
	return target == "x86-64" || target == "rpi"
}

func (d *device) showDownloadTarget(w http.ResponseWriter, r *http.Request) {
	selectedTarget := d.selectedTarget(r.URL.Query())
	sessionId := r.PathValue("sessionId")
	err := d.renderTmpl(w, "device-download-target.tmpl", map[string]any{
		"sessionId":      sessionId,
		"selectedTarget": selectedTarget,
		"wantsWifi":      wantsWifi(selectedTarget),
		"wantsHttpPort":  wantsHttpPort(selectedTarget),
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

func (d *device) showModel(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	template := "model-" + view + ".tmpl"
	if err := d.renderTmpl(w, template, d.Config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
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
func (d *device) editName(w http.ResponseWriter, r *http.Request) {
	if err := d.renderTmpl(w, "edit-name.tmpl", d.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
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
