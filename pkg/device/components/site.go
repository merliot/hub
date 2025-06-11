package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SiteTab struct {
	Name string
	Href string
}

type SiteHeaderParams struct {
	Tabs      []SiteTab
	ActiveTab string
}

// SiteHeader renders the site header with navigation tabs.
func SiteHeader(p SiteHeaderParams) Node {
	return Header(
		Class("flex flex-col mx-4 my-8 max-w-2xl"),
		Div(
			Class("flex flex-row justify-between"),
			A(
				Class("no-underline"),
				Href("/"),
				Div(
					Class("flex flex-col"),
					Span(Class("mr-8 text-purple-200 text-4xl"), Text("MERLIOT")),
					Span(Class("text-purple-200"), Text("DEVICE HUB")),
				),
			),
			Div(
				Class("flex flex-row"),
				Group(
					Map(p.Tabs, func(tab SiteTab) Node {
						if tab.Name == p.ActiveTab {
							return Div(
								Class("flex flex-row items-end justify-end m-0.5 w-28 h-10 bg-yellow-400 border-yellow-400 text-black border-solid border-2 rounded-2xl"),
								Span(Class("mr-2.5 font-bold"), Text(tab.Name)),
							)
						}
						return A(
							Class("no-underline"),
							Href(tab.Href),
							Div(
								Class("flex flex-row items-end justify-end m-0.5 w-20 h-5 bg-purple-200 border-purple-200 text-black border-solid border-2 rounded-xl"),
								Span(Class("mr-2.5 text-sm"), Text(tab.Name)),
							),
						)
					}),
				),
			),
		),
	)
}

// SiteShell renders the main site HTML shell.
type SiteShellParams struct {
	Title      string
	BodyColors string
	Header     Node
	Body       Node
	Footer     Node
}

func SiteShell(p SiteShellParams) Node {
	return El("html",
		Lang("en"),
		Head(
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			Meta(Name("robots"), Content("noindex, nofollow")),
			Meta(Name("referrer"), Content("same-origin")),
			Link(Rel("icon"), Type("image/png"), Attr("sizes", "32x32"), Href("/images/favicon-32x32.png")),
			Link(Rel("icon"), Type("image/png"), Attr("sizes", "16x16"), Href("/images/favicon-16x16.png")),
			Title(p.Title),
			Link(Rel("stylesheet"), Href("/css/tailwind.css")),
			Script(Src("/js/htmx.min.js.gz")),
			Script(Src("/js/htmx-ext-ws.js.gz")),
			Script(Src("/js/util.js")),
		),
		Body(
			Class(p.BodyColors+" m-4"),
			Group([]Node{p.Header, p.Body, p.Footer}),
		),
	)
}

// SiteHome renders the home page content.
func SiteHome(page string, pages []PageTab) Node {
	return Group([]Node{
		Raw(`<style>h3 { color: #ffaa00; }</style>`),
		Div(
			Class("flex flex-row mx-4 my-8"),
			Attr("hx-boost", "true"),
			PageTabs(page, pages),
			Div(
				Class("max-w-lg"),
				Div(
					Attr("hx-get", "/docs/"+page+".html"),
					Attr("hx-trigger", "load"),
					Attr("hx-swap", "outerHTML"),
					Attr("hx-target", "this"),
				),
			),
		),
	})
}

// SiteDemo renders the demo page content.
func SiteDemo(page string, pages []PageTab, session Node) Node {
	return Div(
		Class("flex flex-row mx-4 my-8"),
		Div(
			Class("flex flex-col"),
			PageTabs(page, pages),
			// Add grug quote
			Div(
				Class("mt-8 mr-8 w-36 text-sm"),
				P(Class("italic"), Text(`"working demo especially good trick: force big brain make something to actually work to talk about and code to look at that do thing, will help big brain see reality on ground more quickly"`)),
				A(
					Class("no-underline"),
					Target("_blank"),
					Href("https://grugbrain.dev/"),
					P(Class("text-right text"), Text("-- The Grug Brained Developer")),
				),
			),
		),
		If(page == "devices", Div(Class("mt-4"), session)),
		If(page != "devices", Div(
			Class("max-w-lg"),
			Div(
				Attr("hx-get", "/docs/"+page+".html"),
				Attr("hx-trigger", "load"),
				Attr("hx-swap", "outerHTML"),
				Attr("hx-target", "this"),
			),
		)),
	)
}

// SiteDocs renders the docs page content.
func SiteDocs(page string, pages []PageTab) Node {
	return DeviceDocs(page, pages)
}

// SiteBlog renders the blog page content.
type Blog struct {
	Dir   string
	Title string
	Date  string
}

func SiteBlog(page string, blogs []Blog) Node {
	return Div(
		Class("flex flex-row mx-4 my-8"),
		Div(
			Class("flex flex-col mr-10 items-end text-sm"),
			Div(
				Class("flex flex-row mt-8 w-36 h-10 text-purple-200"),
				Span(Class("font-bold"), Text("BLOGS")),
			),
			Group(
				Map(blogs, func(blog Blog) Node {
					if blog.Dir == page {
						return Div(
							Class("flex flex-col items-end justify-end m-0.5 w-36 h-14 bg-yellow-400 border-yellow-400 text-black text-right border-solid border-2 rounded-2xl"),
							Span(Class("mr-2.5"), Text(blog.Date)),
							Span(Class("mr-2.5 font-bold"), Text(blog.Title)),
						)
					}
					return A(
						Class("no-underline"),
						Href("/blog/"+blog.Dir),
						Div(
							Class("flex flex-col items-end justify-end m-0.5 w-28 bg-purple-200 border-purple-200 text-black text-right border-solid border-2 rounded-xl"),
							Span(Class("mr-2.5"), Text(blog.Title)),
						),
					)
				}),
			),
		),
		Div(
			Class("max-w-lg"),
			Div(
				Attr("hx-get", "/blog/"+page+"/blog.html"),
				Attr("hx-trigger", "load"),
				Attr("hx-swap", "outerHTML"),
				Attr("hx-target", "this"),
			),
		),
	)
}

// DRY helpers for site tabs, header, and shell

func SiteTabsSlice() []SiteTab {
	return []SiteTab{
		{Name: "HOME", Href: "/"},
		{Name: "DEMO", Href: "/demo"},
		{Name: "DOCS", Href: "/doc"},
		{Name: "BLOG", Href: "/blog"},
	}
}

func SiteHeaderDRY(active string) Node {
	return SiteHeader(SiteHeaderParams{Tabs: SiteTabsSlice(), ActiveTab: active})
}

func SiteShellDRY(title string, header, body, footer Node) Node {
	return SiteShell(SiteShellParams{
		Title:      title,
		BodyColors: "bg-black text-purple-200",
		Header:     header,
		Body:       body,
		Footer:     footer,
	})
}
