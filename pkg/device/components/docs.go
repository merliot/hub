package components

import (
	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

type PageTab struct {
	Name  string
	Url   string
	Label string
}

type DeviceDocsParams struct {
	Page  string
	Pages []PageTab
}

// PageTabs renders the documentation page tabs.
func PageTabs(currentPage string, pages []PageTab) Node {
	return html.Div(
		html.Class("flex flex-col mr-10 items-end text-sm"),
		Group(
			Map(pages, func(page PageTab) Node {
				switch {
				case page.Name == "" && page.Url == "":
					return html.Div(
						html.Class("flex flex-row mt-8 w-36 h-10 text-purple-200"),
						html.Span(html.Class("font-bold"), Text(page.Label)),
					)
				case page.Name == currentPage:
					return html.Div(
						html.Class("flex flex-row items-end justify-end m-0.5 w-36 h-10 bg-yellow-400 border-yellow-400 text-black border-solid border-2 rounded-2xl"),
						html.Span(html.Class("mr-2.5 font-bold"), Text(page.Label)),
					)
				default:
					return html.A(
						html.Class("no-underline"),
						html.Href(page.Url),
						html.Div(
							html.Class("flex flex-row items-end justify-end m-0.5 w-32 h-5 bg-purple-200 border-purple-200 text-black border-solid border-2 rounded-xl"),
							html.Span(html.Class("mr-2.5"), Text(page.Label)),
						),
					)
				}
				return nil
			}),
		),
	)
}

// DeviceDocs renders the main documentation page layout.
func DeviceDocs(page string, pages []PageTab) Node {
	return Group([]Node{
		Raw(`<style>h3 { color: #ffaa00; }</style>`),
		html.Div(
			html.Class("flex flex-row mx-4 my-8"),
			Attr("hx-boost", "true"),
			html.Div(
				html.Class("flex flex-col"),
				PageTabs(page, pages),
			),
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
