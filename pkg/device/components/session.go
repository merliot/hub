package components

import (
	"fmt"

	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

type SessionInfo struct {
	ID        string
	Age       string
	LastViews []SessionViewInfo
}

type SessionViewInfo struct {
	DeviceID string
	View     string
	Level    int
}

type SessionsPageParams struct {
	Sessions []SessionInfo
}

// SessionsPage renders the sessions.tmpl view.
func SessionsPage(p SessionsPageParams) Node {
	return El("html",
		html.Lang("en"),
		html.Head(
			html.Meta(html.Name("viewport"), html.Content("width=device-width, initial-scale=1")),
			html.Link(html.Rel("icon"), html.Type("image/png"), Attr("sizes", "32x32"), html.Href("/images/favicon-32x32.png")),
			html.Link(html.Rel("icon"), html.Type("image/png"), Attr("sizes", "16x16"), html.Href("/images/favicon-16x16.png")),
			html.Title("Sessions"),
			html.Link(html.Rel("stylesheet"), html.Href("https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css")),
			html.Meta(Attr("http-equiv", "refresh"), html.Content("2")),
		),
		html.Body(
			html.Class("bg-black text-purple-200 m-4"),
			html.H2(Text("Active Sessions")),
			If(len(p.Sessions) == 0,
				html.P(Text("No active sessions")),
			),
			Group(
				Map(p.Sessions, func(s SessionInfo) Node {
					return html.Div(
						html.Class("bg-yellow-300 text-black m-1 p-2 rounded-3xl max-w-lg"),
						html.H3(Text("Session ID: "+s.ID)),
						html.P(Text("Last Update: "+s.Age)),
						html.H4(Text("Last Views:")),
						html.Ul(
							Group(
								Map(s.LastViews, func(v SessionViewInfo) Node {
									return html.Li(Text(v.DeviceID + ": " + v.View + ", " + itoa(v.Level)))
								}),
							),
						),
					)
				}),
			),
		),
	)
}

type SessionViewParams struct {
	SessionID  string
	PingPeriod int
	Body       Node // Rendered view content
}

// SessionView renders the session.tmpl/session-view.tmpl view.
func SessionView(p SessionViewParams) Node {
	return html.Div(
		html.Class("offline"),
		Attr("id", "session"),
		Attr("hx-headers", `{"session-id": "`+p.SessionID+`"}`),
		Attr("hx-trigger", "every "+itoa(p.PingPeriod)+"s"),
		Attr("ws-send", ""),
		Attr("hx-ext", "ws"),
		Attr("ws-connect", "/wsx?session-id="+p.SessionID),
		p.Body,
	)
}

// itoa is a helper to convert int to string.
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
