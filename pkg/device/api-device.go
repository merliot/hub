//go:build !tinygo

package device

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/merliot/hub/pkg/device/components"
)

func (d *device) installAPI() {

	// Base APIs for all devices

	d.HandleFunc("GET /", d.serveStaticFile)
	d.HandleFunc("GET /{$}", d.showHome)
	d.HandleFunc("GET /show-view", d.showView)
	d.HandleFunc("GET /state", d.showState)
	d.HandleFunc("GET /status", d.showStatus)
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
	case ".go", ".tmpl":
		w.Header().Set("Content-Type", "text/plain")
	case ".gz":
		w.Header().Set("Content-Encoding", "gzip")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	}
	http.FileServer(http.FS(d.layeredFS)).ServeHTTP(w, r)
}

func (d *device) showHome(w http.ResponseWriter, r *http.Request) {
	server := d.server
	sessionId, ok := server.sessions.newSession()
	if !ok {
		server.sessions.noSessions(w, r)
		return
	}
	w.Header().Set("session-id", sessionId)
	// TODO: Replace with gomponents or direct logic for home view
	http.Error(w, "Home view not implemented", http.StatusNotFound)
}

func (d *device) showView(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	sessionId := r.Header.Get("session-id")
	_, level := d.lastView(sessionId)

	if view == "overview" {
		params := components.DeviceOverviewParams{
			Model:        d.Model,
			ClassOffline: "", // Add logic if needed
			ID:           d.Id,
			Level:        level,
			IsOnline:     d.isSet(flagOnline),
			BgColor:      d.Config.BgColor,
			TextColor:    d.Config.FgColor,
			BorderColor:  "", // Add logic if needed
			Name:         d.Name,
			DeployParams: d.DeployParams,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = components.DeviceOverview(params).Render(w)
		return
	}

	if view == "detail" {
		params := components.DeviceDetailParams{
			Model:          d.Model,
			ClassOffline:   "", // Add logic if needed
			ID:             d.Id,
			Level:          level,
			IsOnline:       d.isSet(flagOnline),
			BgColor:        d.Config.BgColor,
			TextColor:      d.Config.FgColor,
			BorderColor:    "", // Add logic if needed
			Name:           d.Name,
			DeployParams:   d.DeployParams,
			SessionID:      sessionId,
			RenderChildren: nil, // Add logic if needed
			Buttons:        nil, // Add logic if needed
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = components.DeviceDetail(params).Render(w)
		return
	}

	if view == "settings" {
		params := components.DeviceSettingsParams{
			Model:          d.Model,
			ClassOffline:   "", // Add logic if needed
			ID:             d.Id,
			Level:          level,
			IsOnline:       d.isSet(flagOnline),
			BgColor:        d.Config.BgColor,
			TextColor:      d.Config.FgColor,
			BorderColor:    "", // Add logic if needed
			Name:           d.Name,
			SessionID:      sessionId,
			RenderChildren: nil, // Add logic if needed
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = components.DeviceSettings(params).Render(w)
		return
	}

	// No legacy rendering fallback. Optionally, return 404 or a default response.
	http.Error(w, "View not found", http.StatusNotFound)
}

func (d *device) showState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(d.stateJSON())
}

func (d *device) showStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(d.statusJSON())
}

func (d *device) showCode(w http.ResponseWriter, r *http.Request) {
	entries, _ := fs.ReadDir(d.layeredFS, ".")
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.Code(names).Render(w)
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
	return target == "pyportal" || target == "wioterminal" ||
		target == "nano-rp2040" || target == "elecrow-rp2040" ||
		target == "elecrow-rp2350"
}

func wantsHttpPort(target string) bool {
	return target == "x86-64" || target == "rpi"
}

func (d *device) showDownloadTarget(w http.ResponseWriter, r *http.Request) {
	selectedTarget := d.selectedTarget(r.URL.Query())
	sessionId := r.PathValue("sessionId")
	params := components.DeviceDownloadTargetParams{
		SessionID:      sessionId,
		SelectedTarget: selectedTarget,
		WantsWifi:      wantsWifi(selectedTarget),
		WantsHttpPort:  wantsHttpPort(selectedTarget),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.DeviceDownloadTarget(params).Render(w)
}

func (d *device) showInstructions(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	// TODO: Load actual instructions content for the view
	content := "Instructions for " + view
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.Instructions(content).Render(w)
}

func (d *device) showInstructionsTarget(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	// TODO: Load actual instructions content for the target
	content := "Instructions for target " + target
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.Instructions(content).Render(w)
}

func (d *device) showModel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.Model(d.Model).Render(w)
}

func (d *device) editName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.EditName(d.Name).Render(w)
}

func (d *device) stateJSON() []byte {
	data, _ := json.Marshal(d.State)
	return data
}
