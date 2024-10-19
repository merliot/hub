package hub

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/websocket"
)

func newConfig(url *url.URL, user, passwd string) (*websocket.Config, error) {
	var surl = url.String()
	var origin = "http://localhost/"

	// Configure the websocket
	config, err := websocket.NewConfig(surl, origin)
	if err != nil {
		return nil, err
	}

	// If valid user, set the basic auth header for the request
	if user != "" {
		req, err := http.NewRequest("GET", surl, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(user, passwd)
		config.Header = req.Header
	}

	return config, nil
}

func wsDial(url *url.URL, user, passwd string) {
	cfg, err := newConfig(url, user, passwd)
	if err != nil {
		fmt.Println("Error configuring websocket:", err)
		return
	}

	for {
		// Dial the websocket
		conn, err := websocket.DialConfig(cfg)
		if err == nil {
			// Service the client websocket
			wsClient(conn)
		} else {
			fmt.Println("Dial error", url, err)
		}

		// Try again in a second
		time.Sleep(time.Second)
	}
}

func wsClient(conn *websocket.Conn) {
	defer conn.Close()

	var link = &wsLink{conn: conn}
	var ann = announcement{
		Id:           root.Id,
		Model:        root.Model,
		Name:         root.Name,
		DeployParams: root.DeployParams,
	}
	var pkt = &Packet{
		Dst:  ann.Id,
		Path: "/announce",
	}

	// Send announcement
	fmt.Println("Sending announment:", pkt)
	err := link.Send(pkt.Marshal(&ann))
	if err != nil {
		fmt.Println("Send error:", err)
		return
	}

	// Receive welcome within 1 sec
	pkt, err = link.receiveTimeout(time.Second)
	if err != nil {
		fmt.Println("Receive error:", err)
		return
	}

	fmt.Println("Reply from announcement:", pkt)
	if pkt.Path != "/welcome" {
		fmt.Println("Not welcomed, got:", pkt.Path)
		return
	}

	fmt.Println("Adding Uplink")
	uplinksAdd(link)

	// Send /state packets to all devices
	devicesSendState(link)

	// Route incoming packets down to the destination device.  Stop and
	// disconnect on EOF.

	fmt.Println("Receiving packets...")
	for {
		pkt, err := link.receivePoll()
		if err != nil {
			fmt.Println("Error receiving packet:", err)
			break
		}
		fmt.Println("Route packet DOWN:", pkt)
		deviceRouteDown(pkt.Dst, pkt)
	}

	fmt.Println("Removing Uplink")
	uplinksRemove(link)
}
