package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
	"github.com/merliot/hub/test/common"
)

var (
	user       = "TEST"
	passwd     = "TESTTEST"
	hubPort    = 8022
	subhubPort = 8023
	sessionId  string
)

var hub = `{
	"gadget1": {
		"Id": "gadget1",
		"Model": "gadget",
		"Name": "Gadget",
		"Children": null,
		"DeployParams": "target=x86-64&port=&Bottles=99"
	},
	"subhub1": {
		"Id": "subhub1",
		"Model": "hub",
		"Name": "Hub",
		"Children": null,
		"DeployParams": "target=x86-64\u0026port=8000"
	},
	"hub": {
		"Id": "hub",
		"Model": "hub",
		"Name": "Hub",
		"Children": [
			"gadget1",
			"subhub1"
		],
		"DeployParams": "target=x86-64\u0026port=8000"
	}
}`

var subhub = `{
	"gadget2": {
		"Id": "gadget2",
		"Model": "gadget",
		"Name": "Gadget",
		"Children": null,
		"DeployParams": "target=x86-64\u0026port=\u0026Bottles=99"
	},
	"subhub1": {
		"Id": "subhub1",
		"Model": "hub",
		"Name": "Hub",
		"Children": [
			"gadget2"
		],
		"DeployParams": "target=x86-64\u0026port=8000"
	}
}`

var merged = `{
	"gadget1": {
		"Id": "gadget1",
		"Model": "gadget",
		"Name": "Gadget",
		"Children": null,
		"DeployParams": "target=x86-64\u0026port=\u0026Bottles=99"
	},
	"gadget2": {
		"Id": "gadget2",
		"Model": "gadget",
		"Name": "Gadget",
		"Children": null,
		"DeployParams": "target=x86-64\u0026port=\u0026Bottles=99"
	},
	"hub": {
		"Id": "hub",
		"Model": "hub",
		"Name": "Hub",
		"Children": [
			"gadget1",
			"subhub1"
		],
		"DeployParams": "target=x86-64\u0026port=8000"
	},
	"subhub1": {
		"Id": "subhub1",
		"Model": "hub",
		"Name": "Hub",
		"Children": [
			"gadget2"
		],
		"DeployParams": "target=x86-64\u0026port=8000"
	}
}`

var gadget2 = `{
	"gadget2": {
		"Id": "gadget2",
		"Model": "gadget",
		"Name": "Gadget",
		"Children": null,
		"DeployParams": "target=x86-64\u0026port=\u0026Bottles=99"
	}
}`

func TestMain(m *testing.M) {

	var err error

	os.WriteFile("devices.json", []byte(hub), 0644)

	// Run a hub
	hubby := device.NewServer(
		device.WithPort(hubPort),
		device.WithModels(models.AllModels),
		//device.WithKeepBuilds("true"),
		device.WithDevicesFile("devices.json"),
		device.WithAutoSave("true"),
		device.WithLogLevel("DEBUG"),
		device.WithDialUrls(",xx://xxx/ws,://example.com"),
		device.WithUser(user),
		device.WithPasswd(passwd),
	)
	go hubby.Run()
	time.Sleep(time.Second)

	// Stash the session id
	sessionId, err = common.GetSession(user, passwd, hubPort)
	if err != nil {
		fmt.Printf("Getting session failed: %s\n", err)
		os.Exit(1)
	}

	code := m.Run()

	os.RemoveAll("devices.json")

	os.Exit(code)
}

func callOK(t *testing.T, port int, method, path string) []byte {
	resp, err := common.CallOK(user, passwd, sessionId, port, method, path)
	if err != nil {
		t.Fatalf("Error %s %s (%d): %s", method, path, port, err)
	}
	return resp
}

func callExpecting(t *testing.T, port int, method, path string, expectedStatus int) []byte {
	resp, err := common.CallExpecting(user, passwd, sessionId, port, method, path, expectedStatus)
	if err != nil {
		t.Fatalf("Error %s %s (%d): %s", method, path, port, err)
	}
	return resp
}

func TestJoin(t *testing.T) {
	// Run a sub-hub
	subby := device.NewServer(
		device.WithPort(subhubPort),
		device.WithModels(models.AllModels),
		device.WithDevicesEnv(subhub),
		device.WithDialUrls("ws://localhost:"+strconv.Itoa(hubPort)+"/ws"),
		device.WithUser(user),
		device.WithPasswd(passwd),
	)
	go subby.Run()
	time.Sleep(time.Second)

	devs := callOK(t, hubPort, "GET", "/devices")
	if !bytes.Equal(devs, []byte(merged)) {
		t.Fatalf("Expected /devices:\n%s\ngot:\n%s", merged, devs)
	}

	// Run gadget2
	models := device.Models{
		"gadget": &models.Gadget,
	}
	g2 := device.NewServer(
		device.WithModels(models),
		device.WithDevicesEnv(gadget2),
		device.WithDialUrls("ws://localhost:"+strconv.Itoa(subhubPort)+"/ws"),
		device.WithUser(user),
		device.WithPasswd(passwd),
	)
	go g2.Run()
	time.Sleep(time.Second)
}

func TestUptime(t *testing.T) {
	callOK(t, hubPort, "POST", "/device/gadget2/get-uptime")
	time.Sleep(time.Second)
}

func TestCreate(t *testing.T) {
	callOK(t, subhubPort, "POST",
		"/create?ParentId=subhub1&Child.Id=test&Child.Model=gadget&Child.Name=test")
	callExpecting(t, subhubPort, "POST",
		"/create?ParentId=subhub1&Child.Id=test&Child.Model=gadget&Child.Name=test", http.StatusBadRequest)
	callExpecting(t, subhubPort, "POST",
		"/create?ParentId=subhub1&Child.Id=test2&Child.Model=XXX&Child.Name=test", http.StatusBadRequest)
}

func TestDownload(t *testing.T) {
	odir, _ := os.Getwd()
	os.Chdir("../../") // Need to chdir to get access to ./bin files
	defer os.Chdir(odir)
	callOK(t, subhubPort, "GET", "/download-image/test?target=x86-64")
	callOK(t, subhubPort, "GET", "/download-image/subhub1?target=x86-64&port=8000")
	callExpecting(t, hubPort, "GET", "/download-image/gadget2?target=x86-64&port=&Bottles=99", http.StatusBadRequest)
}

func TestDestroy(t *testing.T) {
	callExpecting(t, hubPort, "DELETE", "/destroy?Id=test", http.StatusBadRequest)
	callOK(t, subhubPort, "DELETE", "/destroy?Id=test")
}
