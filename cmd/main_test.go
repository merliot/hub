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
	"github.com/stretchr/testify/assert"
)

var (
	user   = "TEST"
	passwd = "TESTTEST"
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
	device.Setenv("DEMO", "true")
	device.Setenv("LOG_LEVEL", "DEBUG")
	//device.Setenv("DEBUG_KEEP_BUILDS", "true")

	device.Models = models.AllModels
	go device.Run()

	time.Sleep(10 * time.Second) // Give the server time to start

	m.Run()
}

func api(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}
	req.SetBasicAuth(user, passwd)
	client := &http.Client{}
	return client.Do(req)
}

func TestRoot(t *testing.T) {

	resp, err := api("GET", "http://localhost:8000/")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
}

func TestAPIDevices(t *testing.T) {

	resp, err := api("GET", "http://localhost:8000/devices")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	html := string(body)

	assert.Equal(t, html, devices, "Comparing devices")
}

func TestAPICreate(t *testing.T) {
	resp, err := api("POST", "http://localhost:8000/create?Id=relaytest&Model=relays&Name=test")
	if err != nil {
		t.Fatalf("API call failed: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
}

func TestAPIDownloadRpi(t *testing.T) {
	odir, _ := os.Getwd()
	os.Chdir("../") // Need to chdir to get access to ./bin files
	defer os.Chdir(odir)

	resp, err := api("GET", "http://localhost:8000/device/relaytest/download-image?target=rpi")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
}

func TestAPIDownloadX86(t *testing.T) {
	odir, _ := os.Getwd()
	os.Chdir("../") // Need to chdir to get access to ./bin files
	defer os.Chdir(odir)

	resp, err := api("GET", "http://localhost:8000/device/relaytest/download-image?target=x86-64")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
}

func TestAPIDownloadNano(t *testing.T) {
	odir, _ := os.Getwd()
	os.Chdir("../") // Need to chdir to get access to ./bin files
	defer os.Chdir(odir)

	resp, err := api("GET", "http://localhost:8000/device/relaytest/download-image?target=nano-rp2040")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
}

func TestAPIDestroy(t *testing.T) {
	resp, err := api("DELETE", "http://localhost:8000/destroy?Id=relaytest")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
}
