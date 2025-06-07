package components

import (
	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
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
	return html.Header(
		html.Class("flex flex-col mx-4 my-8 max-w-2xl"),
		html.Div(
			html.Class("flex flex-row justify-between"),
			html.A(
				html.Class("no-underline"),
				html.Href("/"),
				html.Div(
					html.Class("flex flex-col"),
					html.Span(html.Class("mr-8 text-purple-200 text-4xl"), Text("MERLIOT")),
					html.Span(html.Class("text-purple-200"), Text("DEVICE HUB")),
				),
			),
			html.Div(
				html.Class("flex flex-row"),
				Group(
					Map(p.Tabs, func(tab SiteTab) Node {
						if tab.Name == p.ActiveTab {
							return html.Div(
								html.Class("flex flex-row items-end justify-end m-0.5 w-28 h-10 bg-yellow-400 border-yellow-400 text-black border-solid border-2 rounded-2xl"),
								html.Span(html.Class("mr-2.5 font-bold"), Text(tab.Name)),
							)
						}
						return html.A(
							html.Class("no-underline"),
							html.Href(tab.Href),
							html.Div(
								html.Class("flex flex-row items-end justify-end m-0.5 w-20 h-5 bg-purple-200 border-purple-200 text-black border-solid border-2 rounded-xl"),
								html.Span(html.Class("mr-2.5 text-sm"), Text(tab.Name)),
							),
						)
					})
				),
			),
		),
	)
}

// SiteShell renders the main site HTML shell.
type SiteShellParams struct {
	Title     string
	BodyColors string
	Header    Node
	Body      Node
	Footer    Node
}

func SiteShell(p SiteShellParams) Node {
	return El("html",
		html.Lang("en"),
		html.Head(
			html.Meta(html.Name("viewport"), html.Content("width=device-width, initial-scale=1")),
			html.Meta(html.Name("robots"), html.Content("noindex, nofollow")),
			html.Meta(html.Name("referrer"), html.Content("same-origin")),
			html.Link(html.Rel("icon"), html.Type("image/png"), Attr("sizes", "32x32"), html.Href("/images/favicon-32x32.png")),
			html.Link(html.Rel("icon"), html.Type("image/png"), Attr("sizes", "16x16"), html.Href("/images/favicon-16x16.png")),
			html.Title(p.Title),
			html.Link(html.Rel("stylesheet"), html.Href("https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css")),
			html.Script(html.Src("/js/htmx.min.js.gz")),
			html.Script(html.Src("/js/htmx-ext-ws.js.gz")),
			html.Script(html.Src("/js/util.js")),
		),
		html.Body(
			html.Class(p.BodyColors+" m-4"),
			Group([]Node{p.Header, p.Body, p.Footer}),
		),
	)
}

// SiteHome renders the home page content.
func SiteHome(page string, pages []PageTab) Node {
	return Group([]Node{
		Raw(`<style>h3 { color: #ffaa00; }</style>`),
		html.Div(
			html.Class("flex flex-row mx-4 my-8"),
			Attr("hx-boost", "true"),
			PageTabs(page, pages),
			html.Div(
				html.Class("max-w-lg"),
				html.Div(
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
	return html.Div(
		html.Class("flex flex-row mx-4 my-8"),
		html.Div(
			html.Class("flex flex-col"),
			PageTabs(page, pages),
			// Add grug quote
			html.Div(
				html.Class("mt-8 mr-8 w-36 text-sm"),
				html.P(html.Class("italic"), Text(`"working demo especially good trick: force big brain make something to actually work to talk about and code to look at that do thing, will help big brain see reality on ground more quickly"`)),
				html.A(
					html.Class("no-underline"),
					html.Target("_blank"),
					html.Href("https://grugbrain.dev/"),
					html.P(html.Class("text-right text"), Text("-- The Grug Brained Developer")),
				),
			),
		),
		If(page == "devices", html.Div(html.Class("mt-4"), session)),
		If(page != "devices", html.Div(
			html.Class("max-w-lg"),
			html.Div(
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
	return html.Div(
		html.Class("flex flex-row mx-4 my-8"),
		html.Div(
			html.Class("flex flex-col mr-10 items-end text-sm"),
			html.Div(
				html.Class("flex flex-row mt-8 w-36 h-10 text-purple-200"),
				html.Span(html.Class("font-bold"), Text("BLOGS")),
			),
			Group(
				Map(blogs, func(blog Blog) Node {
					if blog.Dir == page {
						return html.Div(
							html.Class("flex flex-col items-end justify-end m-0.5 w-36 h-14 bg-yellow-400 border-yellow-400 text-black text-right border-solid border-2 rounded-2xl"),
							html.Span(html.Class("mr-2.5"), Text(blog.Date)),
							html.Span(html.Class("mr-2.5 font-bold"), Text(blog.Title)),
						)
					}
					return html.A(
						html.Class("no-underline"),
						html.Href("/blog/"+blog.Dir),
						html.Div(
							html.Class("flex flex-col items-end justify-end m-0.5 w-28 bg-purple-200 border-purple-200 text-black text-right border-solid border-2 rounded-xl"),
							html.Span(html.Class("mr-2.5"), Text(blog.Title)),
						),
					)
				}),
			),
		),
		html.Div(
			html.Class("max-w-lg"),
			html.Div(
				Attr("hx-get", "/blog/"+page+"/blog.html"),
				Attr("hx-trigger", "load"),
				Attr("hx-swap", "outerHTML"),
				Attr("hx-target", "this"),
			),
		),
	)
} 