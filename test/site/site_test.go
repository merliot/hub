package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
)

var (
	user   = "TEST"
	passwd = "TESTTEST"
	addr   = "localhost:8021"
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

	device.Setenv("DEVICES", devices)
	device.Setenv("USER", user)
	device.Setenv("PASSWD", passwd)
	device.Setenv("LOG_LEVEL", "DEBUG")
	//device.Setenv("DEBUG_KEEP_BUILDS", "true")

	// Run a hub in site mode
	device.Setenv("SITE", "true")
	demo := device.NewServer(addr, models.AllModels)
	go demo.Run()
	time.Sleep(time.Second)

	m.Run()

	demo.Stop()

	os.RemoveAll("./camera-images")
}

func callUserPasswd(method, url, user, passwd string) (*http.Response, error) {
	// Little delay so we don't trip the ratelimiter
	time.Sleep(100 * time.Millisecond)

	url = "http://" + addr + url
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}
	req.SetBasicAuth(user, passwd)
	client := &http.Client{}
	return client.Do(req)
}

func call(method, url string) (*http.Response, error) {
	return callUserPasswd(method, url, user, passwd)
}

func callOK(t *testing.T, method, url string) []byte {

	resp, err := call(method, url)
	if err != nil {
		t.Fatalf("API call failed: %v", err)
	}
	defer resp.Body.Close()

	if http.StatusOK != resp.StatusCode {
		t.Fatalf("Expected StatusOK (200), got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return body
}

func callExpecting(t *testing.T, method, url string, expecting int) {

	resp, err := call(method, url)
	if err != nil {
		t.Fatalf("API call failed: %v", err)
	}
	defer resp.Body.Close()

	if expecting != resp.StatusCode {
		t.Fatalf("Expected %d, got %d", expecting, resp.StatusCode)
	}
}

func callBad(t *testing.T, method, url string) {
	callExpecting(t, method, url, http.StatusBadRequest)
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

func TestShowSiteStatus(t *testing.T) {
	callOK(t, "GET", "/status")
	callOK(t, "GET", "/status/devices")
	callOK(t, "GET", "/status/sessions/refresh")
}

func TestShowSiteDocs(t *testing.T) {
	callOK(t, "GET", "/doc")
	callOK(t, "GET", "/doc/faq")
}

func TestShowSiteBlog(t *testing.T) {
	callOK(t, "GET", "/blog")
}
