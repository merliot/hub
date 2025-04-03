package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
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
		"DeployParams": "target=x86-64\u0026port=\u0026Bottles=99"
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
		device.WithLogLevel("DEBUG"),
		device.WithDialUrls(",xx://xxx/ws,://example.com"),
		device.WithUser(user),
		device.WithPasswd(passwd),
	)
	go hubby.Run()
	time.Sleep(time.Second)

	// Stash the session id
	sessionId, err = getSession(hubPort)
	if err != nil {
		fmt.Printf("Getting session failed: %s\n", err)
		os.Exit(1)
	}

	m.Run()
}

var errNoMoreSessions = errors.New("no more sessions")

func getSession(port int) (string, error) {
	// Little delay so we don't trip the ratelimiter
	time.Sleep(100 * time.Millisecond)
	resp, err := call(port, "GET", "/")
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

func callUserPasswd(port int, method, url, user, passwd string) (*http.Response, error) {
	// Little delay so we don't trip the ratelimiter
	time.Sleep(100 * time.Millisecond)

	url = "http://localhost:" + strconv.Itoa(port) + url
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}
	req.SetBasicAuth(user, passwd)
	req.Header.Set("session-id", sessionId)
	client := &http.Client{}
	return client.Do(req)
}

func call(port int, method, url string) (*http.Response, error) {
	return callUserPasswd(port, method, url, user, passwd)
}

func callOK(t *testing.T, port int, method, url string) []byte {

	resp, err := call(port, method, url)
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

func TestSave(t *testing.T) {
	callOK(t, hubPort, "GET", "/save")
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

	callOK(t, hubPort, "POST", "/device/gadget2/get-uptime")
	time.Sleep(time.Second)

	callOK(t, subhubPort, "POST",
		"/create?ParentId=subhub1&Child.Id=test&Child.Model=gadget&Child.Name=test")

	odir, _ := os.Getwd()
	os.Chdir("../../") // Need to chdir to get access to ./bin files
	defer os.Chdir(odir)
	callOK(t, subhubPort, "GET", "/download-image/test?target=x86-64")
	callOK(t, subhubPort, "GET", "/download-image/subhub1?target=x86-64")

	callOK(t, subhubPort, "DELETE", "/destroy?Id=test")

	time.Sleep(time.Second)
}
