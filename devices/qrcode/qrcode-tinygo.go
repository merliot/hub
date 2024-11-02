//go:build tinygo

package qrcode

import (
	"embed"
	"image/color"

	"github.com/merliot/hub"
	goqr "github.com/skip2/go-qrcode"
	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/drivers/ili9341"
)

var fs embed.FS

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
	red   = color.RGBA{255, 0, 0, 255}
	blue  = color.RGBA{0, 0, 255, 255}
	green = color.RGBA{0, 255, 0, 255}
)

type qrcode struct {
	Content string
	display *ili9341.Device
}

func (q *qrcode) GetConfig() hub.Config {
	return hub.Config{
		Model: "qrcode",
		State: q,
		FS:    &fs,
	}
}

func (q *qrcode) Setup() error {
	q.display = initdisplay.InitDisplay()
	width, height := q.display.Size()
	q.display.FillRectangle(0, 0, width/2, height/2, white)
	q.display.FillRectangle(width/2, 0, width/2, height/2, red)
	q.display.FillRectangle(0, height/2, width/2, height/2, green)
	q.display.FillRectangle(width/2, height/2, width/2, height/2, blue)
	q.display.FillRectangle(width/4, height/4, width/2, height/2, black)

	qr, err := goqr.New(q.Content, goqr.Medium)
	if err != nil {
		return err
	}

	bitmap := qr.Bitmap()

	for y, row := range bitmap {
		for x, value := range row {
			if value {
				q.display.SetPixel(int16(x), int16(y), black)
			} else {
				q.display.SetPixel(int16(x), int16(y), white)
			}
		}
	}
	q.display.Display()

	return nil
}
