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
	backup      = dean.GetEnv("BACKUP", "")
	devices     = dean.GetEnv("DEVICES", "")
	demo        = dean.GetEnv("DEMO", "")
	locked      = dean.GetEnv("LOCKED", "")
)

func main() {
	hub := hub.New(id, "hub", name).(*hub.Hub)
	hub.SetWifiAuth(ssids, passphrases)
	hub.SetWsScheme(wsScheme)
	hub.SetBackup(backup)
	hub.SetLocked(locked == "true")
	hub.SetDemo(demo == "true")
	server := dean.NewServer(hub, user, passwd, port)
	hub.SetServer(server)
	hub.RegisterModels()
	hub.LoadDevices(devices)
	server.Run()
}
