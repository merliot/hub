package device

import (
	"net/url"
	"strings"
)

func dialParents(urls string, user, passwd string) {
	for _, u := range strings.Split(urls, ",") {
		if u == "" {
			continue
		}
		url, err := url.Parse(u)
		if err != nil {
			LogError("Parsing URL", "err", err)
			continue
		}
		switch url.Scheme {
		case "ws", "wss":
			go wsDial(url, user, passwd)
		default:
			LogError("Scheme must be ws or wss", "got", u)
		}
	}
}
