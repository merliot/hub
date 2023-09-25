package hub

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

//go:embed *
var fs embed.FS

type Device struct {
	Model  string
	Name   string
	Online bool
}

type Devices map[string]*Device // keyed by id
type Models []string

type Hub struct {
	*common.Common
	Devices
	Models
	server     *dean.Server
	templates  *template.Template
	ssid       string
	passphrase string
	gitKey     string
	gitAuthor  string
}

var targets = []string{"x86-64", "rpi"}

func New(id, model, name string) dean.Thinger {
	println("NEW HUB")
	h := &Hub{}
	h.Common = common.New(id, model, name, targets).(*common.Common)
	h.Devices = make(Devices)
	h.CompositeFs.AddFS(fs)
	h.templates = h.CompositeFs.ParseFS("template/*")
	return h
}

func (h *Hub) SetServer(server *dean.Server) {
	h.server = server
}

func (h *Hub) SetWifiAuth(ssid, passphrase string) {
	h.ssid = ssid
	h.passphrase = passphrase
}

func (h *Hub) SetGit(key, author string) {
	h.gitKey = key
	h.gitAuthor = author
}

func (h *Hub) getState(msg *dean.Msg) {
	h.Path = "state"
	msg.Marshal(h).Reply()
}

func (h *Hub) online(msg *dean.Msg, online bool) {
	var thing dean.ThingMsgConnect
	msg.Unmarshal(&thing)

	if dev, ok := h.Devices[thing.Id]; ok {
		dev.Online = online
	}

	msg.Broadcast()
}

func (h *Hub) connect(online bool) func(*dean.Msg) {
	return func(msg *dean.Msg) {
		h.online(msg, online)
	}
}

func (h *Hub) createdThing(msg *dean.Msg) {
	var create dean.ThingMsgCreated
	msg.Unmarshal(&create)
	h.Devices[create.Id] = &Device{Model: create.Model, Name: create.Name}
	h.storeDevices()
	create.Path = "created/device"
	msg.Marshal(&create).Broadcast()
}

func (h *Hub) deletedThing(msg *dean.Msg) {
	var del dean.ThingMsgDeleted
	msg.Unmarshal(&del)
	delete(h.Devices, del.Id)
	h.storeDevices()
	del.Path = "deleted/device"
	msg.Marshal(&del).Broadcast()
}

func (h *Hub) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"get/state":     h.getState,
		"connected":     h.connect(true),
		"disconnected":  h.connect(false),
		"created/thing": h.createdThing,
		"deleted/thing": h.deletedThing,
	}
}

func (h *Hub) api(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api\n"))
	w.Write([]byte("/create?id={id}&model={model}&name={name}\n"))
	w.Write([]byte("/delete?id={id}\n"))
}

func (h *Hub) apiCreate(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	model := r.URL.Query().Get("model")
	name := r.URL.Query().Get("name")
	thinger, err := h.server.CreateThing(id, model, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	wifiver := thinger.(common.Wifiver)
	wifiver.SetWifiAuth(h.ssid, h.passphrase)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Device id '%s' created", id)
}

func (h *Hub) apiDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := h.server.DeleteThing(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Device id '%s' deleted", id)
}

func (h *Hub) apiSave(w http.ResponseWriter, r *http.Request) {
	if err := h.saveDevices(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Devices saved")
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "api":
		h.api(w, r)
	case "create":
		h.apiCreate(w, r)
	case "delete":
		h.apiDelete(w, r)
	case "save":
		h.apiSave(w, r)
	default:
		h.API(h.templates, w, r)
	}
}

func (h *Hub) restoreDevices() {
	var devices Devices
	bytes, _ := os.ReadFile("devices.json")
	json.Unmarshal(bytes, &devices)
	for id, dev := range devices {
		thinger, err := h.server.CreateThing(id, dev.Model, dev.Name)
		if err != nil {
			fmt.Printf("Skipping: error creating device Id '%s': %s\n", id, err)
			continue
		}
		wifiver := thinger.(common.Wifiver)
		wifiver.SetWifiAuth(h.ssid, h.passphrase)
	}
}

func (h *Hub) storeDevices() {
	bytes, _ := json.MarshalIndent(h.Devices, "", "\t")
	os.WriteFile("devices.json", bytes, 0600)
}

// hasPendingChanges checks if there are any uncommitted changes in the local repo.
func hasPendingChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// commitChanges commits any pending changes in the local repo.
func commitChanges(commitMessage, author string) error {
	commitCmd := exec.Command("git", "commit", "-am", commitMessage, "--author", author)
	out, err := commitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to commit changes: %s, %w", out, err)
	}
	fmt.Println("Changes committed successfully!")
	return nil
}

// pushCommit pushes commits in local repo to remote
func pushCommit(key string) error {
	// 1. Write key to temp file
	tempFile, err := ioutil.TempFile("", "git-ssh-key")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(key)
	if err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tempFile.Close()

	// 2. Set permissions on the temp file
	if err = os.Chmod(tempFile.Name(), 0400); err != nil {
		return fmt.Errorf("failed to set permissions on temp file: %w", err)
	}

	// 3. Set GIT_SSH_COMMAND environment variable
	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no", tempFile.Name())
	os.Setenv("GIT_SSH_COMMAND", sshCmd)

	// 4. Execute git push command
	pushCmd := exec.Command("git", "push")
	out, err := pushCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push commit: %s, %w", out, err)
	}
	fmt.Println("Pushed commit successfully!")
	return nil
}

func (h *Hub) saveDevices() error {
	if h.gitAuthor == "" {
		return errors.New("Can't save: Missing GIT_AUTHOR env var")
	}
	if h.gitKey == "" {
		return errors.New("Can't save: Missing GIT_KEY env var")
	}
	changes, err := hasPendingChanges()
	if err != nil {
		return err
	}
	if !changes {
		return errors.New("No changes to save")
	}
	if err := commitChanges("update devices", h.gitAuthor); err != nil {
		return err
	}
	return pushCommit(h.gitKey)
}

func (h *Hub) dumpDevices() {
	b, _ := json.MarshalIndent(h.Devices, "", "\t")
	fmt.Println(string(b))
}

func (h *Hub) Run(i *dean.Injector) {
	h.Models = h.server.GetModels()
	h.restoreDevices()
	select {}
}
