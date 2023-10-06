//go:build tinygo

package sign

import (
	"encoding/json"
	"image/color"
	"fmt"
	"machine"
	"os"
	"strings"

	"github.com/merliot/dean"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
	"tinygo.org/x/tinyfs/littlefs"
	"tinygo.org/x/tinyterm"
)

const (
	charWidth = 6
	charHeight = 10
)

var (
	black = color.RGBA{0, 0, 0, 255}
	fs    = littlefs.New(machine.Flash)
)

type targetStruct struct {
	display   *ili9341.Device
	terminal  *tinyterm.Terminal
	font      *tinyfont.Font
}

func (s *Sign) targetNew() {
	s.font = &proggy.TinySZ8pt7b
}

func (s *Sign) refresh() {
	lines := strings.Split(s.Banner, "\n")
	for i := 0; i < int(s.Terminal.Height); i++ {
		if i < len(lines) {
			fmt.Fprintln(s.terminal, lines[i])
		} else {
			fmt.Fprintln(s.terminal)
		}
	}
}

func (s *Sign) store() {
	bytes, _ := json.Marshal(s)
	f, err := fs.OpenFile("state", os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		fmt.Fprintln(s.terminal, "error opening file")
		return
	}
	defer f.Close()
	_, err = f.Write(bytes)
	if err != nil {
		fmt.Fprintln(s.terminal, "error writing to file")
	}
}

func (s *Sign) restore() {
	f, err := fs.Open("state")
	if err != nil {
		fmt.Fprintln(s.terminal, "error opening file")
		return
	}
	defer f.Close()
	bytes := make([]byte, 512)
	n, err := f.Read(bytes)
	if err != nil {
		fmt.Fprintln(s.terminal, "error reading file")
		return
	}
	json.Unmarshal(bytes[:n], s)
}

func (s *Sign) mount() {
	if err := fs.Mount(); err != nil {
		// Mount fails on first boot, so format then mount
		if err := fs.Format(); err != nil {
			fmt.Fprintln(s.terminal, "file system format failed")
			return
		} else {
			if err := fs.Mount(); err != nil {
				fmt.Fprintf(s.terminal, "file system mount failed")
				return
			}
		}
	}
}

func (s *Sign) run(i *dean.Injector) {

	// Configure littlefs with parameters for caches and wear levelling
	fs.Configure(&littlefs.Config{
		CacheSize:     512,
		LookaheadSize: 512,
		BlockCycles:   100,
	})

	s.display = initdisplay.InitDisplay()
	s.display.SetRotation(ili9341.Rotation0)

	s.terminal = tinyterm.NewTerminal(s.display)
	s.terminal.Configure(&tinyterm.Config{
		Font:       s.font,
		FontHeight: charHeight,
		FontOffset: charWidth,
	})

	s.mount()
	s.restore()
	s.refresh()

	s.Display.Width, s.Display.Height = s.display.Size()
	s.Terminal.Width = s.Display.Width / charWidth
	s.Terminal.Height = s.Display.Height / charHeight

	select{}
}
