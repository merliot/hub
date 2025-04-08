//go:build !tinygo

package device

import (
	"net/http"
)

type siteTab struct {
	Name string
	Href string
}

type siteTabs []siteTab

var (
	tabHome  = siteTab{"HOME", "/"}
	tabDemo  = siteTab{"DEMO", "/demo"}
	tabDocs  = siteTab{"DOCS", "/doc"}
	tabBlog  = siteTab{"BLOG", "/blog"}
	tabsHome = siteTabs{tabHome, tabDemo, tabDocs, tabBlog}
	tabsDemo = siteTabs{tabDemo, tabHome, tabDocs, tabBlog}
	tabsDocs = siteTabs{tabDocs, tabHome, tabDemo, tabBlog}
	tabsBlog = siteTabs{tabBlog, tabHome, tabDemo, tabDocs}
)

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
