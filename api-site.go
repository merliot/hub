//go:build !tinygo

package hub

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

func (d *device) showSiteHome(w http.ResponseWriter, r *http.Request) {
	d.showSection(w, r, "site.tmpl", "home", "intro", homePages, map[string]any{
		"tabs": tabsHome,
	})
}

var (
	pingPeriod = Getenv("PING_PERIOD", "2")
)

func (d *device) showSiteDemoSession(w http.ResponseWriter, r *http.Request) {
	sessionId, ok := newSession()
	if !ok {
		d.noSessions(w, r)
		return
	}
	d.showSection(w, r, "site.tmpl", "demo", "devices", demoPages, map[string]any{
		"tabs":       tabsDemo,
		"sessionId":  sessionId,
		"pingPeriod": pingPeriod,
	})
}

func (d *device) showSiteDemo(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	if page == "" || page == "devices" {
		d.showSiteDemoSession(w, r)
	} else {
		d.showSection(w, r, "site.tmpl", "demo", "devices", demoPages, map[string]any{
			"tabs": tabsDemo,
		})
	}
}

func (d *device) showSiteStatus(w http.ResponseWriter, r *http.Request) {
	refresh := path.Base(r.URL.Path)
	if refresh == "refresh" {
		d.showStatusRefresh(w, r)
		return
	}
	d.showSection(w, r, "site.tmpl", "status", "sessions", statusPages, map[string]any{
		"tabs":     tabsStatus,
		"sessions": sessionsStatus(),
		"devices":  devicesStatus(),
	})
}

func (d *device) showSiteDocs(w http.ResponseWriter, r *http.Request) {
	d.showSection(w, r, "site.tmpl", "docs", "quick-start", docPages, map[string]any{
		"tabs":   tabsDocs,
		"models": Models,
		"model":  "",
	})
}

func (d *device) showSiteModelDocs(w http.ResponseWriter, r *http.Request) {
	model := r.PathValue("model")
	d.showSection(w, r, "site.tmpl", "docs", "none", docPages, map[string]any{
		"tabs":   tabsDocs,
		"models": Models,
		"model":  model,
	})
}

func (d *device) showSiteBlog(w http.ResponseWriter, r *http.Request) {
	blogs := d.blogs()
	d.showSection(w, r, "site.tmpl", "blog", blogs[0].Dir, nil, map[string]any{
		"tabs":  tabsBlog,
		"blogs": blogs,
	})
}
