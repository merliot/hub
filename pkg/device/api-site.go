//go:build !tinygo

package device

import (
	"net/http"

	"github.com/merliot/hub/pkg/device/components"
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

func (s *server) showSiteHome(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	if page == "" {
		page = "intro"
	}
	tabs := []components.SiteTab{
		{Name: "HOME", Href: "/"},
		{Name: "DEMO", Href: "/demo"},
		{Name: "DOCS", Href: "/doc"},
		{Name: "BLOG", Href: "/blog"},
	}
	pages := make([]components.PageTab, len(homePages))
	for i, p := range homePages {
		pages[i] = components.PageTab{Name: p.Name, Url: p.Url, Label: p.Label}
	}
	header := components.SiteHeader(components.SiteHeaderParams{Tabs: tabs, ActiveTab: "HOME"})
	body := components.SiteHome(page, pages)
	footer := components.SiteFooter()
	site := components.SiteShell(components.SiteShellParams{
		Title:      "Merliot",
		BodyColors: "bg-black text-purple-200",
		Header:     header,
		Body:       body,
		Footer:     footer,
	})
	_ = site.Render(w)
}

func (s *server) showSiteDemo(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	if page == "" {
		page = "devices"
	}
	tabs := []components.SiteTab{
		{Name: "HOME", Href: "/"},
		{Name: "DEMO", Href: "/demo"},
		{Name: "DOCS", Href: "/doc"},
		{Name: "BLOG", Href: "/blog"},
	}
	pages := make([]components.PageTab, len(demoPages))
	for i, p := range demoPages {
		pages[i] = components.PageTab{Name: p.Name, Url: p.Url, Label: p.Label}
	}
	header := components.SiteHeader(components.SiteHeaderParams{Tabs: tabs, ActiveTab: "DEMO"})
	body := components.SiteDemo(page, pages, nil) // TODO: pass session Node
	footer := components.SiteFooter()
	site := components.SiteShell(components.SiteShellParams{
		Title:      "Merliot Demo",
		BodyColors: "bg-black text-purple-200",
		Header:     header,
		Body:       body,
		Footer:     footer,
	})
	_ = site.Render(w)
}

func (s *server) showSiteDocs(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	if page == "" {
		page = "quick-start"
	}
	tabs := []components.SiteTab{
		{Name: "HOME", Href: "/"},
		{Name: "DEMO", Href: "/demo"},
		{Name: "DOCS", Href: "/doc"},
		{Name: "BLOG", Href: "/blog"},
	}
	pages := make([]components.PageTab, len(docPages))
	for i, p := range docPages {
		pages[i] = components.PageTab{Name: p.Name, Url: p.Url, Label: p.Label}
	}
	header := components.SiteHeader(components.SiteHeaderParams{Tabs: tabs, ActiveTab: "DOCS"})
	body := components.SiteDocs(page, pages)
	footer := components.SiteFooter()
	site := components.SiteShell(components.SiteShellParams{
		Title:      "Merliot Docs",
		BodyColors: "bg-black text-purple-200",
		Header:     header,
		Body:       body,
		Footer:     footer,
	})
	_ = site.Render(w)
}

func (s *server) showSiteBlog(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	blogs := s.blogs()
	if page == "" && len(blogs) > 0 {
		page = blogs[0].Dir
	}
	tabs := []components.SiteTab{
		{Name: "HOME", Href: "/"},
		{Name: "DEMO", Href: "/demo"},
		{Name: "DOCS", Href: "/doc"},
		{Name: "BLOG", Href: "/blog"},
	}
	blogTabs := make([]components.Blog, len(blogs))
	for i, b := range blogs {
		blogTabs[i] = components.Blog{Dir: b.Dir, Title: b.Title, Date: b.Date}
	}
	header := components.SiteHeader(components.SiteHeaderParams{Tabs: tabs, ActiveTab: "BLOG"})
	body := components.SiteBlog(page, blogTabs)
	footer := components.SiteFooter()
	site := components.SiteShell(components.SiteShellParams{
		Title:      "Merliot Blog",
		BodyColors: "bg-black text-purple-200",
		Header:     header,
		Body:       body,
		Footer:     footer,
	})
	_ = site.Render(w)
}
