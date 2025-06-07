package components

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SiteFooter renders the sticky site footer.
func SiteFooter() Node {
	return Footer(
		Class("flex flex-row my-4 max-w-2xl items-center w-full fixed bottom-0 bg-transparent"),
		Img(
			Src("/images/made-in-the-usa.png"),
			Class("w-16"),
		),
		Span(
			Class("text-sm ml-4"),
			Text("© 2025 Merliot. All rights reserved."),
		),
	)
}
