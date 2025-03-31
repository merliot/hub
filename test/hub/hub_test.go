package main

import (
	"bytes"
	"errors"
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
	user       = "TEST"
	passwd     = "TESTTEST"
	hubAddr    = "localhost:8022"
	subhubAddr = "localhost:8023"
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

	device.Setenv("DEVICES_FILE", "devices.json")
	device.Setenv("USER", user)
	device.Setenv("PASSWD", passwd)
	device.Setenv("LOG_LEVEL", "DEBUG")
	//device.Setenv("DEBUG_KEEP_BUILDS", "true")

	// Run a hub
	hubby := device.NewServer(hubAddr, models.AllModels)
	go hubby.Run()
	time.Sleep(time.Second)

	// Stash the session id
	sessionId, err = getSession(hubAddr)
	if err != nil {
		fmt.Printf("Getting session failed: %s\n", err)
		os.Exit(1)
	}

	m.Run()

	hubby.Stop()
}

var errNoMoreSessions = errors.New("no more sessions")

func getSession(addr string) (string, error) {
	// Little delay so we don't trip the ratelimiter
	time.Sleep(100 * time.Millisecond)
	resp, err := call(addr, "GET", "/")
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

func callUserPasswd(addr, method, url, user, passwd string) (*http.Response, error) {
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

func call(addr, method, url string) (*http.Response, error) {
	return callUserPasswd(addr, method, url, user, passwd)
}

func callOK(t *testing.T, addr, method, url string) []byte {

	resp, err := call(addr, method, url)
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

func TestJoin(t *testing.T) {
	// Run a sub-hub
	device.Setenv("DEVICES_FILE", "")
	device.Setenv("DEVICES", subhub)
	device.Setenv("DIAL_URLS", "ws://"+hubAddr+"/ws")
	subby := device.NewServer(subhubAddr, models.AllModels)
	go subby.Run()
	time.Sleep(time.Second)

	devs := callOK(t, hubAddr, "GET", "/devices")
	if !bytes.Equal(devs, []byte(merged)) {
		t.Fatalf("Expected /devices:\n%s\ngot:\n%s", merged, devs)
	}

	// Run gadget2
	device.Setenv("DEVICES", gadget2)
	device.Setenv("DIAL_URLS", "ws://"+subhubAddr+"/ws")
	g2 := device.NewServer("", device.Models{
		"gadget": &models.Gadget,
	})
	go g2.Run()
	time.Sleep(time.Second)

	callOK(t, hubAddr, "POST", "/device/gadget2/get-uptime")
	time.Sleep(time.Second)
	subby.Stop()
}
