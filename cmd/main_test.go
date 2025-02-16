package main

import (
	"encoding/json"
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

	device.Models = models.AllModels

	device.Setenv("DEVICES", devices)
	device.Setenv("USER", user)
	device.Setenv("PASSWD", passwd)
	device.Setenv("LOG_LEVEL", "DEBUG")
	//device.Setenv("DEBUG_KEEP_BUILDS", "true")

	// Run hub on :8000 in demo mode
	device.Setenv("PORT", "8000")
	device.Setenv("DEMO", "true")
	go device.Run()

	// Run site on :8001
	device.Setenv("PORT", "8001")
	device.Setenv("SITE", "true")
	go device.Run()

	time.Sleep(time.Second)

	m.Run()

	os.RemoveAll("camera-images")
}

var sessionId string

func api(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}
	req.SetBasicAuth(user, passwd)
	req.Header.Set("session-id", sessionId)
	client := &http.Client{}
	return client.Do(req)
}

func TestRoot(t *testing.T) {

	resp, err := api("GET", "http://localhost:8000/")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	resp.Body.Close()

	sessionId = resp.Header.Get("session-id")
	assert.NotEmpty(t, sessionId)

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

func TestShowViews(t *testing.T) {
	views := []string{"overview", "detail", "settings", "info", "state"}
	devs := make(map[string]any)
	err := json.Unmarshal([]byte(devices), &devs)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for dev, _ := range devs {
		for _, view := range views {
			url := fmt.Sprintf("http://localhost:8000/device/%s/show-view?view=%s", dev, view)
			resp, err := api("GET", url)
			if err != nil {
				t.Fatalf("API failed: %v", err)
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
			resp.Body.Close()
			// Sleep a bit to avoid hitting rate limiter
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func TestAPICreate(t *testing.T) {
	resp, err := api("POST", "http://localhost:8000/create?Id=relaytest&Model=relays&Name=test")
	if err != nil {
		t.Fatalf("API call failed: %v", err)
	}
	resp.Body.Close()

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
	resp.Body.Close()

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
	resp.Body.Close()

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
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
}

func TestAPIDestroy(t *testing.T) {
	resp, err := api("DELETE", "http://localhost:8000/destroy?Id=relaytest")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
}

func TestCamera(t *testing.T) {
	time.Sleep(5 * time.Second)

	resp, err := api("POST", "http://localhost:8000/device/camera1/get-image")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
}

func TestGadget(t *testing.T) {
	resp, err := api("POST", "http://localhost:8000/device/gadget1/takeone")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
	resp.Body.Close()

	time.Sleep(2 * time.Second)
}

func TestQRCode(t *testing.T) {
	resp, err := api("POST", "http://localhost:8000/device/qrcode1/generate?Content=https://foo.com")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
	resp.Body.Close()

	resp, err = api("GET", "http://localhost:8000/device/qrcode1/edit-content?id=qrcode1")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
	resp.Body.Close()
}

func TestRelays(t *testing.T) {
	resp, err := api("POST", "http://localhost:8000/device/relays1/click?Relay=0")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
	resp.Body.Close()

	resp, err = api("POST", "http://localhost:8000/device/relays1/clicked?Relay=1&State=true")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status Code")
	resp.Body.Close()
}
