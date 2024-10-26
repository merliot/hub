//go:build !tinygo

package hub

type page struct {
	// Name of page URL to navigate to.  If Url == "" and Name == "",
	// name means tab heading.
	Name  string
	Label string
	Url   string
}

var homePages = []page{
	page{"", "WELCOME", ""},
	page{"intro", "INTRO", "/home/intro"},
	page{"", "SOURCE", "https://github.com/merliot/hub"},
	page{"contact", "CONTACT", "/home/contact"},
}

var demoPages = []page{
	page{"", "DEMO", ""},
	page{"devices", "DEVICE VIEW", "/demo/devices"},
	page{"network", "NETWORK VIEW", "/demo/network"},
	page{"about-demo", "ABOUT", "/demo/about-demo"},
}

var statusPages = []page{
	page{"", "STATUS", ""},
	page{"sessions", "SESSIONS", "/status/sessions"},
	page{"devices", "DEVICES", "/status/devices"},
}

var docPages = []page{
	page{"", "GUIDES", ""},
	page{"quick-start", "QUICK START", "/doc/quick-start"},
	page{"install", "INSTALL GUIDE", "/doc/install"},
	page{"env-vars", "ENV VARS", "/doc/env-vars"},
	page{"faq", "FAQ", "/doc/faq"},

	page{"", "DEVELOPER", ""},
	page{"", "REFERENCE", "https://pkg.go.dev/github.com/merliot/hub"},
	page{"run-source", "RUN SOURCE", "/doc/run-source"},
	page{"new-model", "NEW MODEL", "/doc/new-model"},
	page{"services", "SERVICES", "/doc/services"},
	page{"api", "API", "/doc/api"},
	page{"ui-device", "UI TO DEVICE", "/doc/ui-device"},
	page{"device-ui", "DEVICE TO UI", "/doc/device-ui"},
	page{"device-views", "DEVICE VIEWS", "/doc/device-views"},
	page{"template-funcs", "TEMPLATE FUNCS", "/doc/template-funcs"},
	page{"template-map", "TEMPLATE MAP", "/doc/template-map"},

	page{"", "DEVICE MODELS", ""},
}
