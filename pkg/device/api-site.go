//go:build !tinygo

package device

import (
	"net/http"

	"github.com/merliot/hub/pkg/device/components"
	"maragu.dev/gomponents"
)

// DRY helper for PageTab-based site pages
func showSitePage(w http.ResponseWriter, r *http.Request, defaultPage, title, activeTab string, pagesSrc []page, bodyFunc func(string, []components.PageTab) gomponents.Node) {
	page := r.PathValue("page")
	if page == "" {
		page = defaultPage
	}
	pages := make([]components.PageTab, len(pagesSrc))
	for i, p := range pagesSrc {
		pages[i] = components.PageTab{Name: p.Name, Url: p.Url, Label: p.Label}
	}
	header := components.SiteHeaderDRY(activeTab)
	body := bodyFunc(page, pages)
	footer := components.SiteFooter()
	site := components.SiteShellDRY(title, header, body, footer)
	if err := site.Render(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) showSiteHome(w http.ResponseWriter, r *http.Request) {
	showSitePage(w, r, "intro", "Merliot", "HOME", homePages, components.SiteHome)
}

func (s *server) showSiteDemo(w http.ResponseWriter, r *http.Request) {
	bodyFunc := func(page string, pages []components.PageTab) gomponents.Node {
		return components.SiteDemo(page, pages, nil) // TODO: pass session Node
	}
	showSitePage(w, r, "devices", "Merliot Demo", "DEMO", demoPages, bodyFunc)
}

func (s *server) showSiteDocs(w http.ResponseWriter, r *http.Request) {
	showSitePage(w, r, "quick-start", "Merliot Docs", "DOCS", docPages, components.SiteDocs)
}

// DRY helper for Blog-based site pages
func showSiteBlogPage(w http.ResponseWriter, r *http.Request, title, activeTab string, blogsSrc []components.Blog, bodyFunc func(string, []components.Blog) gomponents.Node) {
	page := r.PathValue("page")
	if page == "" && len(blogsSrc) > 0 {
		page = blogsSrc[0].Dir
	}
	blogs := make([]components.Blog, len(blogsSrc))
	copy(blogs, blogsSrc)
	header := components.SiteHeaderDRY(activeTab)
	body := bodyFunc(page, blogs)
	footer := components.SiteFooter()
	site := components.SiteShellDRY(title, header, body, footer)
	if err := site.Render(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) showSiteBlog(w http.ResponseWriter, r *http.Request) {
	blogs := s.blogs()
	blogTabs := make([]components.Blog, len(blogs))
	for i, b := range blogs {
		blogTabs[i] = components.Blog{Dir: b.Dir, Title: b.Title, Date: b.Date}
	}
	showSiteBlogPage(w, r, "Merliot Blog", "BLOG", blogTabs, components.SiteBlog)
}
