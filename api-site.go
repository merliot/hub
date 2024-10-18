//go:build !tinygo

package hub

import "net/http"

type siteTab struct {
	Name string
	Href string
}

type siteTabs []siteTab

var (
	tabHome    = siteTab{"HOME", "/home"}
	tabDemo    = siteTab{"DEMO", "/demo"}
	tabStatus  = siteTab{"STATUS", "/status"}
	tabDocs    = siteTab{"DOCS", "/doc"}
	tabBlog    = siteTab{"BLOG", "/blog"}
	tabSource  = siteTab{"SOURCE", "https://github.com/merliot/hub"}
	tabsHome   = siteTabs{tabHome, tabDemo, tabStatus, tabDocs, tabBlog, tabSource}
	tabsDemo   = siteTabs{tabDemo, tabHome, tabStatus, tabDocs, tabBlog, tabSource}
	tabsStatus = siteTabs{tabStatus, tabHome, tabDemo, tabDocs, tabBlog, tabSource}
	tabsDocs   = siteTabs{tabDocs, tabHome, tabDemo, tabStatus, tabBlog, tabSource}
	tabsBlog   = siteTabs{tabBlog, tabHome, tabDemo, tabStatus, tabDocs, tabSource}
)

func (d *device) showSiteHome(w http.ResponseWriter, r *http.Request) {
	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"section": "home",
		"tabs":    tabsHome,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showSiteDemo(w http.ResponseWriter, r *http.Request) {
	sessionId, ok := newSession()
	if !ok {
		d.noSessions(w, r)
		return
	}
	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"sessionId": sessionId,
		"section":   "demo",
		"tabs":      tabsDemo,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showSiteStatus(w http.ResponseWriter, r *http.Request) {
	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"section": "status",
		"tabs":    tabsStatus,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showSiteDocs(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	if page == "" {
		page = "quick-start"
	}
	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"section": "docs",
		"tabs":    tabsDocs,
		"pages":   docPages,
		"page":    page,
		"models":  Models,
		"model":   "",
		"hxget":   "/docs/" + page + ".html",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showSiteModelDocs(w http.ResponseWriter, r *http.Request) {
	model := r.PathValue("model")
	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"section": "docs",
		"tabs":    tabsDocs,
		"pages":   docPages,
		"page":    "",
		"models":  Models,
		"model":   model,
		"hxget":   "/model/" + model + "/docs/doc.html",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *device) showSiteBlog(w http.ResponseWriter, r *http.Request) {
	blogs := d.blogs()
	blog := r.PathValue("blog")
	if blog == "" {
		blog = blogs[0].Dir
	}
	if err := d.renderTmpl(w, "site.tmpl", map[string]any{
		"section": "blog",
		"tabs":    tabsBlog,
		"blogs":   blogs,
		"blog":    blog,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
