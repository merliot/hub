package main

import (
	"flag"
	"fmt"

	"github.com/merliot/dean"
	"github.com/merliot/hub"
)

var (
	id           = dean.GetEnv("ID", "hub01")
	name         = dean.GetEnv("NAME", "Hub")
	deployParams = dean.GetEnv("DEPLOY_PARAMS", "")
	wsScheme     = dean.GetEnv("WS_SCHEME", "ws://")
	port         = dean.GetEnv("PORT", "8000")
	user         = dean.GetEnv("USER", "")
	passwd       = dean.GetEnv("PASSWD", "")
	ssids        = dean.GetEnv("WIFI_SSIDS", "")
	passphrases  = dean.GetEnv("WIFI_PASSPHRASES", "")
	gitRemote    = dean.GetEnv("GIT_REMOTE", "")
	gitKey       = dean.GetEnv("GIT_KEY", "")
	gitAuthor    = dean.GetEnv("GIT_AUTHOR", "")
)

func main() {
	uf2 := flag.Bool("uf2", false, "Generate uf2 files")
	flag.Parse()

	hub := hub.New(id, "hub", name).(*hub.Hub)
	hub.SetDeployParams(deployParams)
	hub.SetWifiAuth(ssids, passphrases)
	hub.SetGit(gitRemote, gitKey, gitAuthor)
	hub.SetWsScheme(wsScheme)
	server := dean.NewServer(hub, user, passwd, port)
	hub.SetServer(server)
	registerModels(hub)

	if *uf2 {
		if err := hub.GenerateUf2s(); err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	server.Run()
}
