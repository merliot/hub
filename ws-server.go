//go:build !tinygo

package hub

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

// ws handles /ws requests
func wsHandle(w http.ResponseWriter, r *http.Request) {
	serv := websocket.Server{Handler: websocket.Handler(wsServer)}
	serv.ServeHTTP(w, r)
}

func wsServer(conn *websocket.Conn) {

	defer conn.Close()

	var link = &wsLink{conn: conn}

	// First receive should be an /announce packet

	pkt, err := link.receive()
	if err != nil {
		fmt.Println("Error receiving first packet:", err)
		return
	}

	if pkt.Path != "/announce" {
		fmt.Println("Not Announcement, got:", pkt.Path)
		return
	}
	fmt.Println("Announcement", pkt)

	var ann announcement
	pkt.Unmarshal(&ann)

	if ann.Id != pkt.Dst {
		fmt.Println("Error: id mismatch", ann.Id, pkt.Dst)
		return
	}

	if ann.Id == root.Id {
		fmt.Println("Error: can't dial into self")
		return
	}

	if err := deviceOnline(ann); err != nil {
		fmt.Println("Device online error:", err)
		return
	}

	// Announcement is good, reply with /welcome packet

	link.Send(pkt.SetPath("/welcome"))

	// Add as active download link

	//fmt.Println("Adding Downlink")
	id := ann.Id
	downlinksAdd(id, link)

	// Route incoming packets up to the destination device.  Stop and
	// disconnect on EOF.

	for {
		pkt, err := link.receivePoll()
		if err != nil {
			fmt.Println("Error receiving packet:", err)
			break
		}
		fmt.Println("Route packet UP:", pkt)
		deviceRouteUp(pkt.Dst, pkt)
	}

	fmt.Println("Removing Downlink")
	downlinksRemove(id)

	deviceOffline(id)
}
