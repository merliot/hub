package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
	"github.com/merliot/hub/test/common"
	"github.com/stretchr/testify/require"
)

var (
	user      = "TEST"
	passwd    = "TESTTEST"
	port      = 8021
	sessionId string
)

var devices = `{
	"camera1": {
		"Id": "camera1",
		"Model": "camera",
		"Name": "Camera",
		"Children": null,
		"DeployParams": "target=rpi\u0026port=8001"
	},
	"gadget1": {
		"Id": "gadget1",
		"Model": "gadget",
		"Name": "Gadget",
		"Children": null,
		"DeployParams": "target=x86-64\u0026port=\u0026Bottles=99"
	},
	"gps1": {
		"Id": "gps1",
		"Model": "gps",
		"Name": "GPS",
		"Children": null,
		"DeployParams": "target=nano-rp2040\u0026Radius=50\u0026PollPeriod=30"
	},
	"hub": {
		"Id": "hub",
		"Model": "hub",
		"Name": "Hub",
		"Children": [
			"gadget1",
			"gps1",
			"locker1",
			"prostar1",
			"qrcode1",
			"relays1",
			"temp1",
			"camera1"
		],
		"DeployParams": "target=x86-64\u0026port=8000"
	},
	"locker1": {
		"Id": "locker1",
		"Model": "locker",
		"Name": "Locker",
		"Children": null,
		"DeployParams": "target=pyportal\u0026Secret=My Secret"
	},
	"prostar1": {
		"Id": "prostar1",
		"Model": "prostar",
		"Name": "Prostar",
		"Children": null,
		"DeployParams": "target=nano-rp2040"
	},
	"qrcode1": {
		"Id": "qrcode1",
		"Model": "qrcode",
		"Name": "QR Code",
		"Children": null,
		"DeployParams": "target=wioterminal\u0026Content=https://merliot.io"
	},
	"relays1": {
		"Id": "relays1",
		"Model": "relays",
		"Name": "Relays",
		"Children": null,
		"DeployParams": "target=nano-rp2040\u0026Relays[0].Name=Lights\u0026Relays[1].Name=Fan #1\u0026Relays[2].Name=Fan #2\u0026Relays[3].Name=\u0026Relays[0].Gpio=D2\u0026Relays[1].Gpio=D3\u0026Relays[2].Gpio=D4\u0026Relays[3].Gpio="
	},
	"temp1": {
		"Id": "temp1",
		"Model": "temp",
		"Name": "Temp/Hum",
		"Children": null,
		"DeployParams": "target=nano-rp2040\u0026Sensor=BME280\u0026TempUnits=F\u0026Gpio="
	}
}`

func TestMain(m *testing.M) {
	// Run a hub in site mode
	demo := device.NewServer(
		device.WithPort(port),
		device.WithModels(models.AllModels),
		//device.WithKeepBuilds("true"),
		device.WithRunningSite("true"),
		device.WithDevicesEnv(devices),
		device.WithLogLevel("DEBUG"),
		device.WithUser(user),
		device.WithPasswd(passwd),
	)
	go demo.Run()
	time.Sleep(time.Second)

	// Stash the session id
	var err error
	sessionId, err = common.GetSession(user, passwd, port)
	if err != nil {
		fmt.Printf("Getting session failed: %s\n", err)
		os.Exit(1)
	}

	code := m.Run()

	err = os.RemoveAll("./camera-images")
	if err != nil {
		fmt.Printf("Error removing camera-images: %s\n", err)
	}

	os.Exit(code)
}

func callOK(t *testing.T, method, path string) []byte {
	resp, err := common.CallOK(user, passwd, sessionId, port, method, path)
	require.NoError(t, err, "Error %s %s (%d): %s", method, path, port, err)
	return resp
}

func TestShowSiteHome(t *testing.T) {
	callOK(t, "GET", "/")
	callOK(t, "GET", "/home")
	callOK(t, "GET", "/home/contact")
}

func TestShowSiteDemo(t *testing.T) {
	callOK(t, "GET", "/demo")
	callOK(t, "GET", "/demo/devices")
	callOK(t, "GET", "/demo/about-demo")
}

func TestShowSiteDocs(t *testing.T) {
	callOK(t, "GET", "/doc")
	callOK(t, "GET", "/doc/faq")
}

func TestShowSiteBlog(t *testing.T) {
	callOK(t, "GET", "/blog")
}
