package camera

import (
	"embed"
	"fmt"
	"html/template"
	"time"

	"github.com/merliot/hub/devices/camera/cache"
	"github.com/merliot/hub/pkg/device"
)

//go:embed *.go images template
var embedFS embed.FS

type camera struct {
	*cache.Cache
}

type msgGetImage struct {
	Index uint32
}

type msgImage struct {
	Jpeg []byte
	Prev uint32
	Next uint32
}

func NewModel() device.Devicer {
	return &camera{Cache: cache.New()}
}

func (c *camera) GetConfig() device.Config {
	return device.Config{
		Model:      "camera",
		State:      c,
		FS:         &embedFS,
		Targets:    []string{"rpi", "x86-64"},
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
	msgImage.Jpeg, msgImage.Prev, msgImage.Next, err = c.Cache.GetJpeg(msgGet.Index)
	if err == nil {
		pkt.SetPath("/image").Marshal(&msgImage).RouteUp()
	} else {
		println(err.Error())
	}
}

func (c *camera) jpeg(raw string) (template.URL, error) {
	url := fmt.Sprintf("data:image/jpeg;base64,%s", raw)
	// Return it as template-safe url to use with <img src={{.}}>
	return template.URL(url), nil
}

func (c *camera) Setup() error {
	return c.Cache.Init()
}

func (c *camera) poll() {

	// Capture the jpeg image from the webcam
	jpeg, err := captureJpeg()
	if err != nil {
		fmt.Printf("Error capturing image: %v\n", err)
		return
	}

	// Save the jpeg image to the cache (and to disc)
	err = c.Cache.SaveJpeg(jpeg)
	if err != nil {
		fmt.Printf("Error saving image: %v\n", err)
		return
	}
}

func (c *camera) Poll(pkt *device.Packet) {
	// Run image capture/save in separate go func so device lock is not
	// held long during Polling
	go c.poll()
}
