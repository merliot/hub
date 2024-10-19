package hub

import (
	"fmt"
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
			fmt.Printf("Error parsing URL: %s\r\n", err.Error())
			continue
		}
		switch url.Scheme {
		case "ws", "wss":
			go wsDial(url, user, passwd)
		default:
			fmt.Println("Scheme must be ws or wss:", u)
		}
	}
}
