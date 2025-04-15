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
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
	"github.com/merliot/hub/test/common"
)

var (
	user      = "TEST"
	passwd    = "TESTTEST"
	port      = 8020
	sessionId string
)

var devices = `{
	"timer1": {
		"Id": "timer1",
		"Model": "timer",
		"Name": "Timer",
		"Children": null,
		"DeployParams": "target=nano-rp2040\u0026StartHHMM=15:00\u0026StopHHMM=16:00"
	},
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
			"timer1",
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

	// Run a hub in demo mode
	demo := device.NewServer(
		device.WithPort(port),
		device.WithModels(models.AllModels),
		//device.WithKeepBuilds("true"),
		device.WithRunningDemo("true"),
		device.WithBackground("GOOD"),
		device.WithWifiSsids("foo"),
		device.WithWifiPassphrases("bar"),
		device.WithDevicesEnv(devices),
		device.WithLogLevel("DEBUG"),
		device.WithUser(user),
		device.WithPasswd(passwd),
	)
	go demo.Run()
	time.Sleep(time.Second)

	// Stash the session id
	sessionId, err = common.GetSession(user, passwd, port)
	if err != nil {
		fmt.Printf("Getting session failed: %s\n", err)
		os.Exit(1)
	}

	// Simulate a browser session by opening a /wsx websocket
	url := "ws://localhost:" + strconv.Itoa(port) + "/wsx?session-id=" + sessionId
	go wsx(url, user, passwd)

	code := m.Run()

	os.RemoveAll("./camera-images")
	os.RemoveAll("./raw-1.jpg")

	os.Exit(code)
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

func callOK(t *testing.T, method, path string) []byte {
	resp, err := common.CallOK(user, passwd, sessionId, port, method, path)
	if err != nil {
		t.Fatalf("Error %s %s (%d): %s", method, path, port, err)
	}
	return resp
}

func callBad(t *testing.T, method, path string) []byte {
	return callExpecting(t, method, path, http.StatusBadRequest)
}

func callExpecting(t *testing.T, method, path string, expectedStatus int) []byte {
	resp, err := common.CallExpecting(user, passwd, sessionId, port, method, path, expectedStatus)
	if err != nil {
		t.Fatalf("Error %s %s (%d): %s", method, path, port, err)
	}
	return resp
}

func TestMaxSessions(t *testing.T) {
	for i := 0; i < 99; i++ {
		_, err := common.GetSession(user, passwd, port)
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err := common.GetSession(user, passwd, port)
	if err != common.ErrNoMoreSessions {
		t.Fatal("Expected no more sessions")
	}
}

func TestBadUser(t *testing.T) {
	_, err := common.CallExpecting("bad", "user", sessionId, port,
		"GET", "/", http.StatusUnauthorized)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
}

func devsEqual(a, b string) bool {
	type device struct {
		Id           string
		Model        string
		Name         string
		DeployParams string
		Children     []string
	}
	type devices map[string]device
	var aa = make(devices)
	var bb = make(devices)
	if err := json.Unmarshal([]byte(a), &aa); err != nil {
		println(err.Error())
	}
	if err := json.Unmarshal([]byte(b), &bb); err != nil {
		println(err.Error())
	}
	return reflect.DeepEqual(aa, bb)
}

func TestAPIDevices(t *testing.T) {
	devs := callOK(t, "GET", "/devices")
	if !devsEqual(string(devs), devices) {
		t.Fatalf("/devices response not valid, got: %s\nwant: %s\n", devs, devices)
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

	for dev := range devs {
		for _, view := range views {
			url := fmt.Sprintf("/device/%s/show-view?view=%s", dev, view.name)
			callExpecting(t, "GET", url, view.expectedStatus)
		}
	}
}

var expectedGadgetState = []byte(`{
	"Bottles": 99,
	"Restock": 70,
	"FullCount": 99
}`)

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
	callOK(t, "GET", "/device/qrcode1/instructions?view=collapsed")
}

func TestShowInstructionsTarget(t *testing.T) {
	callOK(t, "GET", "/device/qrcode1/instructions-target?target=x86-64")
}

func TestShowModel(t *testing.T) {
	callOK(t, "GET", "/model/gps/model?view=collapsed")
}

func TestEditName(t *testing.T) {
	callOK(t, "GET", "/device/gps1/edit-name")
	callExpecting(t, "GET", "/device/XXX/edit-name", http.StatusNotFound)
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
	callOK(t, "GET", "/download-image/relaytest?target=nano-rp2040&ssid=foo")
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

func TestGhost(t *testing.T) {
	callExpecting(t, "GET", "/device/relaytest/state", http.StatusGone)
}

func TestAPIRecreate(t *testing.T) {
	callOK(t, "POST",
		"/create?ParentId=hub&Child.Id=relaytest&Child.Model=relays&Child.Name=test")
}

func TestNOP(t *testing.T) {
	callOK(t, "PUT", "/nop")
}

func TestSave(t *testing.T) {
	callOK(t, "POST", "/save")
}

func TestSaveModal(t *testing.T) {
	callOK(t, "GET", "/save-modal")
}

func TestRename(t *testing.T) {
	callOK(t, "PUT", "/rename?Id=relays1&NewName=foo")
	callBad(t, "PUT", "/rename?Id=relays1&NewName=")
	callBad(t, "PUT", "/rename?Id=XXX&NewName=foo")
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
}
