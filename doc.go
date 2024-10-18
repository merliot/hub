//go:build !tinygo

package hub

type docPage struct {
	Name  string
	Label string
}

var docPages = []docPage{
	docPage{"", "GETTING STARTED"},
	docPage{"quick-start", "QUICK START"},
	docPage{"install", "INSTALL GUIDE"},
	docPage{"env-vars", "ENV VARS"},
	docPage{"faqs", "FAQS"},
	docPage{"", "DEVELOPER"},
	docPage{"new-model", "NEW MODEL"},
	docPage{"services", "SERVICES"},
	docPage{"api", "API"},
	docPage{"device-views", "DEVICE VIEWS"},
	docPage{"template-funcs", "TEMPLATE FUNCS"},
	docPage{"template-map", "TEMPLATE MAP"},
	docPage{"", "DEVICE MODELS"},
}
