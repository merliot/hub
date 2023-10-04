//go:build wioterminal

package sign

import (
	"image/color"
	"fmt"
	"machine"

	"github.com/merliot/dean"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
	"tinygo.org/x/tinyterm"
)

var (
	black = color.RGBA{0, 0, 0, 255}
)

type targetStruct struct {
	display   *ili9341.Device
	backlight machine.Pin
	terminal  *tinyterm.Terminal
	font      *tinyfont.Font
}

func (s *Sign) targetNew() {
	s.display = ili9341.NewSPI(
		machine.SPI3,
		machine.LCD_DC,
		machine.LCD_SS_PIN,
		machine.LCD_RESET,
	)
	s.backlight = machine.LCD_BACKLIGHT
	s.terminal = tinyterm.NewTerminal(s.display)
	s.font = &proggy.TinySZ8pt7b
}

func (s *Sign) run(i *dean.Injector) {

	machine.SPI3.Configure(machine.SPIConfig{
		SCK:       machine.LCD_SCK_PIN,
		SDO:       machine.LCD_SDO_PIN,
		SDI:       machine.LCD_SDI_PIN,
		Frequency: 40000000,
	})

	s.display.Configure(ili9341.Config{})
	s.display.FillScreen(black)

	s.backlight.Configure(machine.PinConfig{machine.PinOutput})
	s.backlight.High()

	s.terminal.Configure(&tinyterm.Config{
		Font:       s.font,
		FontHeight: 10,
		FontOffset: 6,
	})

	fmt.Fprintf(s.terminal, "Hello, World!")

	select{}
}
