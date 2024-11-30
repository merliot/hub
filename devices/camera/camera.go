package camera

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/merliot/hub/pkg/device"
)

//go:embed *.go images template
var embedFS embed.FS

type camera struct {
	getJpeg func(int) ([]byte, error)
	latest  string
	sync.RWMutex
}

type msgGetImage struct {
	Index int
}

type msgImage struct {
	Jpeg []byte
}

func NewModel() device.Devicer {
	return &camera{}
}

func (c *camera) GetConfig() device.Config {
	return device.Config{
		Model:      "camera",
		State:      c,
		FS:         &embedFS,
		Targets:    []string{"rpi"},
		BgColor:    "almond-creme",
		FgColor:    "black",
		PollPeriod: 5 * time.Second,
		PacketHandlers: device.PacketHandlers{
			"/get-image": &device.PacketHandler[msgGetImage]{c.getImage},
			"/image":     &device.PacketHandler[msgImage]{device.RouteUp},
		},
		FuncMap: template.FuncMap{
			"jpeg": c.jpeg,
		},
	}
}

func (c *camera) getImage(pkt *device.Packet) {
	var msgGet msgGetImage
	var msgImage msgImage
	var err error

	pkt.Unmarshal(&msgGet)
	msgImage.Jpeg, err = c.getJpeg(msgGet.Index)
	if err == nil {
		pkt.SetPath("/image").Marshal(&msgImage).RouteUp()
	} else {
		println(err.Error())
	}
}

func (c *camera) jpeg(raw string) (template.URL, error) {
	// Convert the image to base64
	//base64Image := base64.StdEncoding.EncodeToString([]byte(raw))
	url := fmt.Sprintf("data:image/jpeg;base64,%s", raw)
	// Return it as template-safe url to use with <img src={{.}}>
	return template.URL(url), nil
}

func (c *camera) rawJpeg(index int) ([]byte, error) {
	c.RLock()
	defer c.RUnlock()

	if c.latest == "" {
		return nil, fmt.Errorf("Image file not set yet")
	}

	file, err := os.Open(c.latest)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file contents
	return io.ReadAll(file)
}

func (c *camera) Setup() error {
	c.getJpeg = c.rawJpeg
	return nil
}

func (c *camera) poll() {

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("image_%s.jpg", timestamp)

	cmd := exec.Command("libcamera-still", "-o", filename, "--width", "640", "--height", "480", "-t", "1", "--immediate")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error capturing image: %v\n", err)
	} else {
		fmt.Printf("Captured %s\n", filename)
		c.Lock()
		c.latest = filename
		c.Unlock()
	}
}

func (c *camera) Poll(pkt *device.Packet) {
	go c.poll()
}
