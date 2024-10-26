//go:build !tinygo

package hub

type page struct {
	Name  string // empty means tab heading
	Label string
}

var docPages = []page{
	page{"", "GETTING STARTED"},
	page{"quick-start", "QUICK START"},
	page{"install", "INSTALL GUIDE"},
	page{"env-vars", "ENV VARS"},
	page{"faq", "FAQ"},
	page{"", "DEVELOPER"},
	page{"new-model", "NEW MODEL"},
	page{"services", "SERVICES"},
	page{"api", "API"},
	page{"ui-device", "UI TO DEVICE"},
	page{"device-ui", "DEVICE TO UI"},
	page{"device-views", "DEVICE VIEWS"},
	page{"template-funcs", "TEMPLATE FUNCS"},
	page{"template-map", "TEMPLATE MAP"},
	page{"", "DEVICE MODELS"},
}

var statusPages = []page{
	page{"", "STATUS"},
	page{"sessions", "SESSIONS"},
	page{"devices", "DEVICES"},
}
