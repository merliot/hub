package device

import (
	"net/url"
	"strings"
)

func (s *server) dialParents(urls string, user, passwd string) {
	for _, u := range strings.Split(urls, ",") {
		if u == "" {
			continue
		}
		url, err := url.Parse(u)
		if err != nil {
			s.LogError("Parsing URL", "err", err)
			continue
		}
		switch url.Scheme {
		case "ws", "wss":
			go s.wsDial(url, user, passwd)
		default:
			s.LogError("Scheme must be ws or wss", "got", u)
		}
	}
}
