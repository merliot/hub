//go:build tinygo

package move

import (
	"machine"
	"time"

	"github.com/merliot/dean"
	"tinygo.org/x/drivers/lsm6dsox"
)

type targetStruct struct {
}

func (m *Move) targetNew() {
}

func calibrateGyro(device *lsm6dsox.Device) bool {
	var cal = [...]float32{0, 0, 0}
	for i := 0; i < 100; i++ {
		gx, gy, gz, _ := device.ReadRotation()
		cal[0] += float32(gx) / 1000000
		cal[1] += float32(gy) / 1000000
		cal[2] += float32(gz) / 1000000
		time.Sleep(time.Millisecond * 10)
	}
	cal[0] /= 100
	cal[1] /= 100
	cal[2] /= 100
	// heuristic: after successful calibration the value can't be 0
	return cal[0] != 0
}

func (m *Move) Run(i *dean.Injector) {

	var msg dean.Msg
	var update = Update{Path: "update"}

	machine.I2C0.Configure(machine.I2CConfig{})
	device := lsm6dsox.New(machine.I2C0)
	if err := device.Configure(lsm6dsox.Configuration{
		AccelRange:      lsm6dsox.ACCEL_2G,
		AccelSampleRate: lsm6dsox.ACCEL_SR_104,
		GyroRange:       lsm6dsox.GYRO_250DPS,
		GyroSampleRate:  lsm6dsox.GYRO_SR_104,
	}); err != nil {
		println("LSM6DSOX could not be configured")
		return
	}

	println("LSM6DSOX calibrating...")

	i.Inject(msg.Marshal(CALIBRATE_BEGIN))

	for !calibrateGyro(device) {
		time.Sleep(time.Second)
	}

	i.Inject(msg.Marshal(CALIBRATE_END))

	println("LSM6DSOX calibrating done")

	for {
		update.Ax, update.Ay, update.Az, _ = device.ReadAcceleration()
		update.Gx, update.Gy, update.Gz, _ = device.ReadRotation()
		i.Inject(msg.Marshal(update))
		time.Sleep(time.Second)
	}
}
