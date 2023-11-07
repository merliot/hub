package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/gps"
	"github.com/merliot/hub/models/hub"
	"github.com/merliot/hub/models/ps30m"
	"github.com/merliot/hub/models/hp2430n"
	"github.com/merliot/hub/models/relays"
	"github.com/merliot/hub/models/sign"
	"github.com/merliot/hub/models/uv"
	"github.com/merliot/hub/models/move"
	"github.com/merliot/hub/models/skeleton"
)

func main() {
	h := hub.New("hub01", "hub", "hub01").(*hub.Hub)

	h.ParseWifiAuth()

	gitRemote := os.Getenv("GIT_REMOTE")
	gitKey := os.Getenv("GIT_KEY")
	gitAuthor := os.Getenv("GIT_AUTHOR")
	h.SetGit(gitRemote, gitKey, gitAuthor)

	_, h.BackupHub = os.LookupEnv("BACKUP_HUB")

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
	server.RegisterModel("hp2430n", hp2430n.New)
	server.RegisterModel("gps", gps.New)
	server.RegisterModel("relays", relays.New)
	server.RegisterModel("sign", sign.New)
	server.RegisterModel("uv", uv.New)
	server.RegisterModel("move", move.New)
	server.RegisterModel("skeleton", skeleton.New)
	//server.RegisterModel("hub", hub.New)

	if port, ok := os.LookupEnv("PPROF_PORT"); ok {
		go func() {
			log.Println(http.ListenAndServe(":"+port, nil))
		}()
	}

	go server.ListenAndServe()
	server.Run()
}
