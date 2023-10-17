//go:build tinygo

package uv

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	"tinygo.org/x/drivers/veml6070"
)

type targetStruct struct {
}

func (u *Uv) targetNew() {
}

func (u *Uv) Run(i *dean.Injector) {

	var msg dean.Msg
	var update = Update{Path: "update"}

	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := veml6070.New(machine.I2C0)

	if !sensor.Configure() {
		println("VEML6070 could not be configured")
		return
	}

	println("VEML6070 configured")

	for {
		intensitymW, _ := sensor.ReadUVALightIntensity()
		intensityW := float32(intensitymW) / 1000.0
		riskLevel := RiskLevel(sensor.GetEstimatedRiskLevel(intensitymW))
		if intensityW != u.Intensity {
			update.Intensity, update.RiskLevel = intensityW, riskLevel
			i.Inject(msg.Marshal(update))
		}
		time.Sleep(time.Second)
	}
}
