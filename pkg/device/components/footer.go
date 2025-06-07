package components

import (
	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

// SiteFooter renders the sticky site footer.
func SiteFooter() Node {
	return html.Footer(
		html.Class("flex flex-row my-4 max-w-2xl items-center w-full fixed bottom-0 bg-transparent"),
		html.Img(
			html.Src("/images/made-in-the-usa.png"),
			html.Class("w-16"),
		),
		html.Span(
			html.Class("text-sm ml-4"),
			Text("© 2025 Merliot. All rights reserved."),
		),
	)
}
