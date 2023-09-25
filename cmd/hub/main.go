package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/gps"
	"github.com/merliot/hub/models/hub"
	"github.com/merliot/hub/models/ps30m"
	"github.com/merliot/hub/models/relays"
)

func main() {
	h := hub.New("hub01", "hub", "hub01").(*hub.Hub)

	ssid := os.Getenv("SSID")
	passphrase := os.Getenv("PASSPHRASE")
	h.SetWifiAuth(ssid, passphrase)

	gitKey := os.Getenv("GIT_KEY")
	gitAuthor := os.Getenv("GIT_AUTHOR")
	h.SetGit(gitKey, gitAuthor)

	server := dean.NewServer(h)
	h.SetServer(server)

	server.Addr = ":8000"
	if port, ok := os.LookupEnv("PORT"); ok {
		server.Addr = ":" + port
	}

	if user, ok := os.LookupEnv("USER"); ok {
		if passwd, ok := os.LookupEnv("PASSWD"); ok {
			server.BasicAuth(user, passwd)
		}
	}

	server.RegisterModel("ps30m", ps30m.New)
	server.RegisterModel("gps", gps.New)
	server.RegisterModel("relays", relays.New)
	//server.RegisterModel("hub", hub.New)

	go server.ListenAndServe()
	server.Run()
}
