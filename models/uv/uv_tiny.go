//go:build tinygo

package uv

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	"tinygo.org/x/drivers/veml6070"
)

func (u *UV) Run(i *dean.Injector) {

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
		intensity, _ := sensor.ReadUVALightIntensity()
		riskLevel := RiskLevel(sensor.GetEstimatedRiskLevel(intensity))
		if intensity != u.Intensity {
			update.Intensity, update.RiskLevel = intensity, riskLevel
			i.Inject(msg.Marshal(update))
		}
		time.Sleep(time.Second)
	}
}
