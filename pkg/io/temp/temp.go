//go:build !tinygo && !rpi

package temp

type Temp struct {
	Temperature float32
	Humidity    float32
	Sensor      string `schema:"desc=Sensor name"`
	Gpio        string `schema:"desc=GPIO pin"`
}

func (t *Temp) Setup(sensor, gpio string) error                  { return nil }
func (t *Temp) Read() (temperature, humidity float32, err error) { return 0, 0, nil }
