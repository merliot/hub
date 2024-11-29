//go:build tinygo

package qrcode

import (
	"fmt"
	"image/color"
	"machine"
	"os"

	"github.com/merliot/hub/pkg/device"
	goqr "github.com/skip2/go-qrcode"
	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/tinyfs/littlefs"
)

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
)

type qrcode struct {
	Content string
	display *ili9341.Device
	lfs     *littlefs.LFS
}

func (q *qrcode) GetConfig() device.Config {
	return device.Config{
		Model: "qrcode",
		State: q,
		PacketHandlers: device.PacketHandlers{
			"/update": &device.PacketHandler[qrcode]{q.update},
		},
	}
}

func (q *qrcode) update(pkt *device.Packet) {
	pkt.Unmarshal(q)
	if err := q.paint(); err == nil {
		// save q.Content to FLASH
		q.writeContent()
		pkt.BroadcastUp()
	}
}

func (q *qrcode) paint() error {

	q.display.FillScreen(white)

	content := q.Content
	if content == "" {
		content = "missing content?"
	}

	qr, err := goqr.New(content, goqr.Medium)
	if err != nil {
		return err
	}

	bitmap := qr.Bitmap()
	bw := int16(len(bitmap))
	bh := bw

	if bw == 0 {
		return fmt.Errorf("zero length qr code bitmap")
	}

	// scale bitmap to fit fully within display
	dw, dh := q.display.Size()
	scale := dw / bw
	if scale > (dh / bh) {
		scale = dh / bh
	}

	// actual width, height
	w := bw * scale
	h := bh * scale

	// offset from 0,0
	ow := (dw - w - scale) / 2
	oh := (dh - h - scale) / 2

	for r, row := range bitmap {
		for c, value := range row {
			x := (int16(c) * scale) + ow
			y := (int16(r) * scale) + oh
			if value {
				q.display.FillRectangle(x, y, scale, scale, black)
			}
		}
	}

	return nil
}

// setupFS sets up the FLASH file system
func (q *qrcode) setupFS() error {
	q.lfs = littlefs.New(machine.Flash)

	// Configure littlefs with parameters for caches and wear levelling
	q.lfs.Configure(&littlefs.Config{
		CacheSize:     512,
		LookaheadSize: 512,
		BlockCycles:   100,
	})

	// Try to mount the filesystem
	err := q.lfs.Mount()
	if err != nil {
		device.LogInfo("Filesystem not formatted. Formatting now...")
		// If mounting fails, format the filesystem
		err = q.lfs.Format()
		if err != nil {
			return fmt.Errorf("Failed to format filesystem: %w", err)
		}

		// Mount again after formatting
		err = q.lfs.Mount()
		if err != nil {
			return fmt.Errorf("Failed to mount after formatting: %w", err)
		}
	}

	return nil
}

var file = "/content"

// readContent reads q.Content from fs
func (q *qrcode) readContent() {

	readFile, err := q.lfs.Open(file)
	if err != nil {
		return
	}
	defer readFile.Close()

	buf := make([]byte, 512)
	n, err := readFile.Read(buf)
	if err != nil {
		return
	}

	q.Content = string(buf[:n])
}

// writeContent writes q.Content to fs
func (q *qrcode) writeContent() {

	writeFile, err := q.lfs.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return
	}
	defer writeFile.Close()

	_, err = writeFile.Write([]byte(q.Content))
	if err != nil {
		return
	}
}

func (q *qrcode) Setup() error {
	q.display = initdisplay.InitDisplay()
	if err := q.setupFS(); err != nil {
		return err
	}
	// Read q.Content from FLASH, if previously stored
	q.readContent()
	return q.paint()
}
