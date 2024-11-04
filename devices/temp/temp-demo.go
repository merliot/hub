package temp

import (
	"math"
	"math/rand"
	"time"

	"github.com/merliot/hub"
)

func (t *temp) DemoSetup() error {
	rand.Seed(time.Now().UnixNano())
	return nil
}

func randomValue(mean, stddev float64) float32 {
	u1 := rand.Float64()
	u2 := rand.Float64()

	// Box-Muller transform to generate normally distributed random numbers
	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
	value := mean + stddev*z0

	// Round to 1 dec place
	return float32(math.Round(value*10) / 10)
}

func (t *temp) DemoPoll(pkt *hub.Packet) {
	var msg = msgUpdate{
		Temperature: randomValue(24.1, 0.05),
		Humidity:    randomValue(34.5, 0.05),
	}
	t.Temperature = msg.Temperature
	t.Humidity = msg.Humidity
	pkt.SetPath("/update").Marshal(&msg).RouteUp()
}
