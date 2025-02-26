//go:build !tinygo

package device

import (
	"net/http"
	"path"
)

type siteTab struct {
	Name string
	Href string
}

type siteTabs []siteTab

var (
	tabHome    = siteTab{"HOME", "/"}
	tabDemo    = siteTab{"DEMO", "/demo"}
	tabStatus  = siteTab{"STATUS", "/status"}
	tabDocs    = siteTab{"DOCS", "/doc"}
	tabBlog    = siteTab{"BLOG", "/blog"}
	tabsHome   = siteTabs{tabHome, tabDemo, tabStatus, tabDocs, tabBlog}
	tabsDemo   = siteTabs{tabDemo, tabHome, tabStatus, tabDocs, tabBlog}
	tabsStatus = siteTabs{tabStatus, tabHome, tabDemo, tabDocs, tabBlog}
	tabsDocs   = siteTabs{tabDocs, tabHome, tabDemo, tabStatus, tabBlog}
	tabsBlog   = siteTabs{tabBlog, tabHome, tabDemo, tabStatus, tabDocs}
)

func (d *device) setupSiteAPI() {
	d.installSiteAPIs()
	d.packetHandlersInstall()
}

func (d *device) installSiteAPIs() {

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

func (s *server) showSiteHome(w http.ResponseWriter, r *http.Request) {
	s.root.showSection(w, r, "site.tmpl", "home", "intro", homePages, map[string]any{
		"tabs": tabsHome,
	})
}

func (s *server) showSiteDemoSession(w http.ResponseWriter, r *http.Request) {
	sessionId, ok := s.sessions.newSession()
	if !ok {
		s.sessions.noSessions(w, r)
		return
	}
	s.root.showSection(w, r, "site.tmpl", "demo", "devices", demoPages, map[string]any{
		"tabs":       tabsDemo,
		"sessionId":  sessionId,
		"pingPeriod": s.wsxPingPeriod,
	})
}

func (s *server) showSiteDemo(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	if page == "" || page == "devices" {
		s.showSiteDemoSession(w, r)
	} else {
		s.root.showSection(w, r, "site.tmpl", "demo", "devices", demoPages, map[string]any{
			"tabs": tabsDemo,
		})
	}
}

func (s *server) showSiteStatus(w http.ResponseWriter, r *http.Request) {
	refresh := path.Base(r.URL.Path)
	if refresh == "refresh" {
		s.showStatusRefresh(w, r)
		return
	}
	s.root.showSection(w, r, "site.tmpl", "status", "sessions", statusPages, map[string]any{
		"tabs":     tabsStatus,
		"sessions": s.sessions.status(),
		"devices":  s.devices.status(),
	})
}

func (s *server) showSiteDocs(w http.ResponseWriter, r *http.Request) {
	s.root.showSection(w, r, "site.tmpl", "docs", "quick-start", docPages, map[string]any{
		"tabs": tabsDocs,
	})
}

func (s *server) showSiteBlog(w http.ResponseWriter, r *http.Request) {
	blogs := s.blogs()
	s.root.showSection(w, r, "site.tmpl", "blog", blogs[0].Dir, nil, map[string]any{
		"tabs":  tabsBlog,
		"blogs": blogs,
	})
}
