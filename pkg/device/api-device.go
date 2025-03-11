//go:build !tinygo

package device

import (
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
)

type APIs map[string]http.HandlerFunc

func (d *device) installAPI() {

	// Base APIs for all devices

	d.HandleFunc("GET /", d.serveStaticFile)
	d.HandleFunc("GET /show-view", d.showView)
	d.HandleFunc("GET /state", d.showState)
	d.HandleFunc("GET /code", d.showCode)
	d.HandleFunc("GET /download-target/{sessionId}", d.showDownloadTarget)
	d.HandleFunc("GET /instructions", d.showInstructions)
	d.HandleFunc("GET /instructions-target", d.showInstructionsTarget)
	d.HandleFunc("GET /edit-name", d.editName)
	d.HandleFunc("GET /model", d.showModel)

	// Device-specific APIs, if any

	for path, fn := range d.APIs {
		d.HandleFunc(path, fn)
	}
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

func (d *device) editName(w http.ResponseWriter, r *http.Request) {
	if err := d.renderTmpl(w, "edit-name.tmpl", d.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
