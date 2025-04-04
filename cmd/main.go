// Merliot Hub
//
// To run standalone:
//   $ go run ./cmd
//
// To generate x84-64 and rpi binaries, run:
//   $ go generate ./cmd

//go:generate sh -x -c "go run ./gen-models/ ../models.json ../pkg/models/models.go"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o ../bin/device-x86-64 -tags x86_64 ./"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GOARM=5 go build -ldflags '-s -w' -o ../bin/device-rpi -tags rpi ./"

package main

import (
	"os"
	"strconv"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
)

func getenv(name string, defaultValue string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	return value
}

func main() {

	port, _ := strconv.Atoi(getenv("PORT", "8000"))
	pingPeriod, _ := strconv.Atoi(getenv("PING_PERIOD", "2"))
	keepBuilds := getenv("DEBUG_KEEP_BUILDS", "")
	runningSite := getenv("SITE", "")
	runningDemo := getenv("DEMO", "")
	background := getenv("BACKGROUND", "")
	wifiSsids := getenv("WIFI_SSIDS", "")
	wifiPassphrases := getenv("WIFI_PASSPHRASES", "")
	autoSave := getenv("AUTO_SAVE", "true")
	devicesEnv := getenv("DEVICES", "")
	devicesFile := getenv("DEVICES_FILE", "")
	logLevel := getenv("LOG_LEVEL", "INFO")
	dialUrls := getenv("DIAL_URLS", "")
	user := getenv("USER", "")
	passwd := getenv("PASSWD", "")

	server := device.NewServer(
		device.WithPort(port),
		device.WithModels(models.AllModels),
		device.WithPingPeriod(pingPeriod),
		device.WithKeepBuilds(keepBuilds),
		device.WithRunningSite(runningSite),
		device.WithRunningDemo(runningDemo),
		device.WithBackground(background),
		device.WithWifiSsids(wifiSsids),
		device.WithWifiPassphrases(wifiPassphrases),
		device.WithAutoSave(autoSave),
		device.WithDevicesEnv(devicesEnv),
		device.WithDevicesFile(devicesFile),
		device.WithLogLevel(logLevel),
		device.WithDialUrls(dialUrls),
		device.WithUser(user),
		device.WithPasswd(passwd),
	)

	server.Run()
}
