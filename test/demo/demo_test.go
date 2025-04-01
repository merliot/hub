package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
)

var (
	user      = "TEST"
	passwd    = "TESTTEST"
	addr      = "localhost:8020"
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

	var err error

	device.Setenv("WIFI_SSIDS", "foo")
	device.Setenv("WIFI_PASSPHRASES", "bar")
	device.Setenv("BACKGROUND", "GOOD")
	device.Setenv("DEVICES", devices)
	device.Setenv("USER", user)
	device.Setenv("PASSWD", passwd)
	device.Setenv("LOG_LEVEL", "DEBUG")
	//device.Setenv("DEBUG_KEEP_BUILDS", "true")

	// Run a hub in demo mode
	device.Setenv("DEMO", "true")
	demo := device.NewServer(addr, models.AllModels)
	go demo.Run()
	time.Sleep(time.Second)

	// Stash the session id
	sessionId, err = getSession()
	if err != nil {
		fmt.Printf("Getting session failed: %s\n", err)
		os.Exit(1)
	}

	// Simulate a browser session by opening a /wsx websocket
	url := "ws://" + addr + "/wsx?session-id=" + sessionId
	go wsx(url, user, passwd)

	m.Run()

	demo.Stop()

	os.RemoveAll("./camera-images")
	os.RemoveAll("./raw-1.jpg")
}

var errNoMoreSessions = errors.New("no more sessions")

func wsx(url, user, passwd string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(user, passwd)
	conn, _, err := websocket.DefaultDialer.Dial(url, req.Header)
	if err != nil {
		println(err.Error())
	}
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			println(err.Error())
		}
	}
}

func getSession() (string, error) {
	// Little delay so we don't trip the ratelimiter
	time.Sleep(100 * time.Millisecond)
	resp, err := call("GET", "/")
	if err != nil {
		return "", fmt.Errorf("API failed: %v\n", err)
	}
	resp.Body.Close()
	if http.StatusTooManyRequests == resp.StatusCode {
		return "", errNoMoreSessions
	}
	if http.StatusOK != resp.StatusCode {
		return "", fmt.Errorf("Bad HTTP Status Code %d", resp.StatusCode)
	}
	sessionId := resp.Header.Get("session-id")
	return sessionId, nil
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
	req.Header.Set("session-id", sessionId)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if http.StatusOK != resp.StatusCode {
		println(string(body))
		t.Fatalf("Expected StatusOK (200), got %d", resp.StatusCode)
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

func callNotFound(t *testing.T, method, url string) {
	callExpecting(t, method, url, http.StatusNotFound)
}

func TestMaxSessions(t *testing.T) {
	for i := 0; i < 99; i++ {
		_, err := getSession()
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err := getSession()
	if err != errNoMoreSessions {
		t.Fatal("Expected no more sessions")
	}
}

func TestBadUser(t *testing.T) {
	resp, err := callUserPasswd("GET", "/", "foo", "bar")
	if err != nil {
		t.Fatalf("API failed: %v", err)
	}
	resp.Body.Close()
	if http.StatusUnauthorized != resp.StatusCode {
		t.Fatalf("Bad HTTP Status Code %d", resp.StatusCode)
	}
}

func TestAPIDevices(t *testing.T) {
	html := string(callOK(t, "GET", "/devices"))
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

		file, err := os.ReadFile("../../pkg/device/" + fileName)
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

		body := callOK(t, "GET", "/"+fileName)
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
		{"overview", http.StatusOK},
	}
	devs := make(map[string]any)
	err := json.Unmarshal([]byte(devices), &devs)
	if err != nil {
		t.Fatal(err)
	}

	for dev, _ := range devs {
		for _, view := range views {
			url := fmt.Sprintf("/device/%s/show-view?view=%s", dev, view.name)
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
	state := callOK(t, "GET", "/device/gadget1/state")
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
	code := callOK(t, "GET", "/device/gps1/code")
	if !bytes.Equal(expectedGpsCode, code) {
		t.Fatalf("/code mismatch:\nexpected: %s\napi: %s",
			string(expectedGpsCode), string(code))
	}
}

func TestDownloadTarget(t *testing.T) {
	callOK(t, "GET", "/device/locker1/download-target/xxx")
}

func TestShowInstructions(t *testing.T) {
	callOK(t, "GET", "/device/qrcode1/instructions?view=collasped")
}

func TestShowInstructionsTarget(t *testing.T) {
	callOK(t, "GET", "/device/qrcode1/instructions-target?target=x86-64")
}

func TestShowModel(t *testing.T) {
	callOK(t, "GET", "/model/gps/model?view=collasped")
}

func TestEditName(t *testing.T) {
	callOK(t, "GET", "/device/gps1/edit-name")
	callNotFound(t, "GET", "/device/XXX/edit-name")
}

func TestAPICreate(t *testing.T) {
	callOK(t, "POST",
		"/create?ParentId=hub&Child.Id=relaytest&Child.Model=relays&Child.Name=test")
	callBad(t, "POST",
		"/create?ParentId=XXX&Child.Id=relaytest&Child.Model=relays&Child.Name=test")
	callBad(t, "POST",
		"/create?ParentId=hub&Child.Id=&Child.Model=relays&Child.Name=test")
	callBad(t, "POST",
		"/create?ParentId=hub&Child.Id=-xxx&Child.Model=relays&Child.Name=test")
	callBad(t, "POST",
		"/create?ParentId=hub&Child.Id=x,x&Child.Model=relays&Child.Name=test")
	callBad(t, "POST",
		"/create?ParentId=hub&Child.Id=testtest&Child.Model=XXX&Child.Name=test")
	callBad(t, "POST",
		"/create?ParentId=hub&Child.Id=relaytest&Child.Model=relays&Child.Name=")
}

func TestAPIDownload(t *testing.T) {
	odir, _ := os.Getwd()
	os.Chdir("../../") // Need to chdir to get access to ./bin files
	defer os.Chdir(odir)
	callOK(t, "GET", "/download-image/relaytest?target=rpi")
	callOK(t, "GET", "/download-image/relaytest?target=x86-64")
	callOK(t, "GET", "/download-image/relaytest?target=nano-rp2040")
	callOK(t, "GET", "/deploy-koyeb/relaytest/"+sessionId)
	callBad(t, "GET", "/deploy-koyeb/XXX/"+sessionId)
	callBad(t, "GET", "/download-image/relaytest?target=XXX")
	callBad(t, "GET", "/download-image/XXX?target=rpi")
}

func TestAPIDestroy(t *testing.T) {
	callOK(t, "DELETE", "/destroy?Id=relaytest")
	callBad(t, "DELETE", "/destroy?Id=hub")
	callBad(t, "DELETE", "/destroy?Id=XXX")
	callBad(t, "DELETE", "/destroy")
}

func TestAPIRecreate(t *testing.T) {
	callOK(t, "POST",
		"/create?ParentId=hub&Child.Id=relaytest&Child.Model=relays&Child.Name=test")
}

func TestSave(t *testing.T) {
	callOK(t, "GET", "/save")
}

func TestSaveModal(t *testing.T) {
	callOK(t, "GET", "/save-modal")
}

func TestRename(t *testing.T) {
	callOK(t, "GET", "/rename?Id=relays1&NewName=foo")
	callBad(t, "GET", "/rename?Id=relays1&NewName=")
	callBad(t, "GET", "/rename?Id=XXX&NewName=foo")
}

func TestNewModal(t *testing.T) {
	callOK(t, "GET", "/new-modal/hub")
	callOK(t, "GET", "/new-modal/relays1")
	callBad(t, "GET", "/new-modal/XXX")
}

func TestCamera(t *testing.T) {
	time.Sleep(5 * time.Second)
	callOK(t, "POST", "/device/camera1/get-image")
}

func TestGadget(t *testing.T) {
	callOK(t, "POST", "/device/gadget1/takeone")
	time.Sleep(2 * time.Second)
	callOK(t, "POST", "/device/gadget1/reboot")
}

func TestQRCode(t *testing.T) {
	callOK(t, "POST", "/device/qrcode1/generate?Content=https://foo.com")
	callOK(t, "GET", "/device/qrcode1/edit-content?id=qrcode1")
}

func TestRelays(t *testing.T) {
	callOK(t, "POST", "/device/relays1/click?Relay=0")
	callOK(t, "POST", "/device/relays1/clicked?Relay=1&State=true")
}
