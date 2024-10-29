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
	{"", "WELCOME", ""},
	{"intro", "INTRO", "/home/intro"},
	{"", "SOURCE", "https://github.com/merliot/hub"},
	{"contact", "CONTACT", "/home/contact"},
}

var demoPages = []page{
	{"", "DEMO", ""},
	{"devices", "DEVICE VIEW", "/demo/devices"},
	{"network", "NETWORK VIEW", "/demo/network"},
	{"about-demo", "ABOUT", "/demo/about-demo"},
}

var statusPages = []page{
	{"", "STATUS", ""},
	{"sessions", "SESSIONS", "/status/sessions"},
	{"devices", "DEVICES", "/status/devices"},
}

var docPages = []page{
	{"", "GUIDES", ""},
	{"quick-start", "QUICK START", "/doc/quick-start"},
	{"install", "INSTALL GUIDE", "/doc/install"},
	{"env-vars", "ENV VARS", "/doc/env-vars"},
	{"faq", "FAQ", "/doc/faq"},

	{"", "DEVELOPER", ""},
	{"", "REFERENCE", "https://pkg.go.dev/github.com/merliot/hub"},
	{"run-source", "RUN SOURCE", "/doc/run-source"},
	{"new-model", "NEW MODEL", "/doc/new-model"},
	{"services", "SERVICES", "/doc/services"},
	{"api", "API", "/doc/api"},
	{"ui-device", "UI TO DEVICE", "/doc/ui-device"},
	{"device-ui", "DEVICE TO UI", "/doc/device-ui"},
	{"device-views", "DEVICE VIEWS", "/doc/device-views"},
	{"template-funcs", "TEMPLATE FUNCS", "/doc/template-funcs"},
	{"template-map", "TEMPLATE MAP", "/doc/template-map"},

	{"", "DEVICE MODELS", ""},
}
