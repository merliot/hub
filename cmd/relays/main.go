package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/relays"
)

func main() {
	thing := relays.New("relays01", "relays", "relays").(*relays.Relays)

	thing.SetDeployParams("target=rpi&amp;http=on&amp;relay1=&amp;relay2=&amp;relay3=&amp;relay4=&amp;gpio1=GPIO06&amp;gpio2=GPIO13&amp;gpio3=GPIO19&amp;gpio4=GPIO26")

	server := dean.NewServer(thing)

	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("USER")
	passwd, _ := os.LookupEnv("PASSWD")

	server.DialWebSocket(user, passwd, "ws://192.168.1.213:8000/ws/1500", thing.Announce())

	if port != "" {
		server.Addr = ":" + port
		go server.ListenAndServe()
	}

	server.Run()
}
