package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub"
)

var (
	id          = dean.GetEnv("ID", "hub01")
	name        = dean.GetEnv("NAME", "Hub")
	wsScheme    = dean.GetEnv("WS_SCHEME", "ws://")
	port        = dean.GetEnv("PORT", "8000")
	user        = dean.GetEnv("USER", "")
	passwd      = dean.GetEnv("PASSWD", "")
	ssids       = dean.GetEnv("WIFI_SSIDS", "")
	passphrases = dean.GetEnv("WIFI_PASSPHRASES", "")
	dialURLs    = dean.GetEnv("DIAL_URLS", "")
	devices     = dean.GetEnv("DEVICES", "")
)

func main() {
	hub := hub.New(id, "hub", name).(*hub.Hub)
	hub.SetWifiAuth(ssids, passphrases)
	hub.SetWsScheme(wsScheme)
	hub.SetDialURLs(dialURLs)
	server := dean.NewServer(hub, user, passwd, port)
	hub.SetServer(server)
	hub.RegisterModels()
	hub.LoadDevices(devices)
	server.Run()
}
