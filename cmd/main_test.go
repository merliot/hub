package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
)

var (
	user          = "TEST"
	passwd        = "TESTTEST"
	demoAddr      = "localhost:8020"
	siteAddr      = "localhost:8021"
	demoSessionId string
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

	// Run hub in demo mode
	device.Setenv("DEMO", "true")
	demo := device.NewServer(demoAddr, models.AllModels)
	go demo.Run()
	time.Sleep(time.Second)

	// Stash the session id
	demoSessionId = getSession()

	// Run hub in site mode
	device.Setenv("SITE", "true")
	site := device.NewServer(siteAddr, models.AllModels)
	go site.Run()
	time.Sleep(time.Second)

	m.Run()

	os.RemoveAll("camera-images")
}

func getSession() string {
	resp, err := call("GET", "http://"+demoAddr+"/")
	if err != nil {
		fmt.Printf("API failed: %v\n", err)
		os.Exit(1)
	}
	resp.Body.Close()
	sessionId := resp.Header.Get("session-id")
	if sessionId == "" {
		fmt.Println("No session ID returned")
		os.Exit(1)
	}
	if http.StatusOK != resp.StatusCode {
		fmt.Printf("Bad HTTP Status Code %d", resp.StatusCode)
		os.Exit(1)
	}
	return sessionId
}

func callUserPasswd(method, url, user, passwd string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}
	req.SetBasicAuth(user, passwd)
	req.Header.Set("session-id", demoSessionId)
	client := &http.Client{}
	return client.Do(req)
}

func call(method, url string) (*http.Response, error) {
	return callUserPasswd(method, url, user, passwd)
}

func testCallOK(t *testing.T, method, url string) []byte {

	resp, err := call(method, url)
	if err != nil {
		t.Fatalf("API call failed: %v", err)
	}
	defer resp.Body.Close()

	if http.StatusOK != resp.StatusCode {
		t.Fatalf("Bad HTTP Status Code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return body
}

func TestBadUser(t *testing.T) {
	resp, err := callUserPasswd("GET", "http://"+demoAddr+"/", "foo", "bar")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	resp.Body.Close()
	if http.StatusUnauthorized != resp.StatusCode {
		t.Fatalf("Bad HTTP Status Code %d", resp.StatusCode)
	}
}

func TestAPIDevices(t *testing.T) {
	html := string(testCallOK(t, "GET", "http://"+demoAddr+"/devices"))
	if html != devices {
		t.Fatalf("/devices response not valid")
	}
}

func decompressGzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func TestFiles(t *testing.T) {

	var files = []string{
		"robots.txt",
		"template/device.tmpl",
		"css/device.css.gz",
		"js/util.js",
	}

	for _, fileName := range files {

		file, err := os.ReadFile("../pkg/device/" + fileName)
		if err != nil {
			t.Fatal(err)
		}

		if strings.HasSuffix(fileName, ".gz") {
			var err error
			file, err = decompressGzip(file)
			if err != nil {
				t.Fatal(err)
			}
		}

		body := testCallOK(t, "GET", "http://"+demoAddr+"/"+fileName)
		if !bytes.Equal(file, body) {
			t.Fatalf("Content mismatch:\nfile: %s\napi: %s",
				string(file), string(body))
		}
	}
}

func TestShowViews(t *testing.T) {
	type view struct {
		name           string
		expectedStatus int
	}
	views := []view{
		{"overview", http.StatusOK},
		{"detail", http.StatusOK},
		{"settings", http.StatusOK},
		{"info", http.StatusOK},
		{"state", http.StatusOK},
		{"garbage", http.StatusBadRequest},
	}
	devs := make(map[string]any)
	err := json.Unmarshal([]byte(devices), &devs)
	if err != nil {
		t.Fatal(err)
	}

	for dev, _ := range devs {
		for _, view := range views {
			url := fmt.Sprintf("http://%s/device/%s/show-view?view=%s",
				demoAddr, dev, view.name)
			resp, err := call("GET", url)
			if err != nil {
				t.Fatalf("API failed: %v", err)
			}
			if view.expectedStatus != resp.StatusCode {
				t.Fatalf("Unexpected HTTP status code, got: %d want: %d",
					resp.StatusCode, view.expectedStatus)
			}
			resp.Body.Close()
			// Sleep a bit to avoid hitting rate limiter
			time.Sleep(100 * time.Millisecond)
		}
	}
}

var expectedGadgetState = []byte(`<pre class="text-sm"
	hx-get="/device/gadget1/state"
	hx-trigger="load delay:1s"
	hx-target="this"
	hx-swap="outerHTML">{
	&#34;Bottles&#34;: 99,
	&#34;Restock&#34;: 70,
	&#34;FullCount&#34;: 99
}</pre>
`)

func TestShowState(t *testing.T) {
	state := testCallOK(t, "GET", "http://"+demoAddr+"/device/gadget1/state")
	if !bytes.Equal(expectedGadgetState, state) {
		t.Fatalf("/state mismatch:\nexpected: %s\napi: %s",
			string(expectedGadgetState), string(state))
	}
}

var expectedGpsCode = []byte(`<!DOCTYPE html>
<html>
  <body>
    <pre>
<a href="gps-demo.go">gps-demo.go</a>
<a href="gps-linux.go">gps-linux.go</a>
<a href="gps-tinygo.go">gps-tinygo.go</a>
<a href="gps.go">gps.go</a>
<a href="images">images</a>
<a href="template">template</a>

    </pre>
  </body>
</html>
`)

func TestShowCode(t *testing.T) {
	code := testCallOK(t, "GET", "http://"+demoAddr+"/device/gps1/code")
	if !bytes.Equal(expectedGpsCode, code) {
		t.Fatalf("/code mismatch:\nexpected: %s\napi: %s",
			string(expectedGpsCode), string(code))
	}
}

func TestDownloadTarget(t *testing.T) {
	testCallOK(t, "GET", "http://"+demoAddr+"/device/locker1/download-target/xxx")
}

func TestShowInstructions(t *testing.T) {
	testCallOK(t, "GET", "http://"+demoAddr+"/device/qrcode1/instructions?view=collasped")
}

func TestShowInstructionsTarget(t *testing.T) {
	testCallOK(t, "GET", "http://"+demoAddr+"/device/qrcode1/instructions-target?target=x86-64")
}

func TestShowModel(t *testing.T) {
	testCallOK(t, "GET", "http://"+demoAddr+"/model/gps/model?view=collasped")
}

func TestEditName(t *testing.T) {
	testCallOK(t, "GET", "http://"+demoAddr+"/device/gps1/edit-name")
}

func TestAPICreate(t *testing.T) {
	testCallOK(t, "POST", "http://"+demoAddr+
		"/create?ParentId=hub&Child.Id=relaytest&Child.Model=relays&Child.Name=test")
}

func TestAPIDownloadRpi(t *testing.T) {
	odir, _ := os.Getwd()
	os.Chdir("../") // Need to chdir to get access to ./bin files
	defer os.Chdir(odir)
	testCallOK(t, "GET", "http://"+demoAddr+"/download-image/relaytest?target=rpi")
}

func TestAPIDownloadX86(t *testing.T) {
	odir, _ := os.Getwd()
	os.Chdir("../") // Need to chdir to get access to ./bin files
	defer os.Chdir(odir)
	testCallOK(t, "GET", "http://"+demoAddr+"/download-image/relaytest?target=x86-64")
}

func TestAPIDownloadNano(t *testing.T) {
	odir, _ := os.Getwd()
	os.Chdir("../") // Need to chdir to get access to ./bin files
	defer os.Chdir(odir)
	testCallOK(t, "GET", "http://"+demoAddr+"/download-image/relaytest?target=nano-rp2040")
}

func TestAPIDestroy(t *testing.T) {
	testCallOK(t, "DELETE", "http://"+demoAddr+"/destroy?Id=relaytest")
}

func TestCamera(t *testing.T) {
	time.Sleep(5 * time.Second)
	testCallOK(t, "POST", "http://"+demoAddr+"/device/camera1/get-image")
}

func TestGadget(t *testing.T) {
	testCallOK(t, "POST", "http://"+demoAddr+"/device/gadget1/takeone")
	time.Sleep(2 * time.Second)
}

func TestQRCode(t *testing.T) {
	testCallOK(t, "POST", "http://"+demoAddr+"/device/qrcode1/generate?Content=https://foo.com")
	testCallOK(t, "GET", "http://"+demoAddr+"/device/qrcode1/edit-content?id=qrcode1")
}

func TestRelays(t *testing.T) {
	testCallOK(t, "POST", "http://"+demoAddr+"/device/relays1/click?Relay=0")
	testCallOK(t, "POST", "http://"+demoAddr+"/device/relays1/clicked?Relay=1&State=true")
}

func TestMaxSessions(t *testing.T) {
	for i := 0; i < 100; i++ {
	}
}
