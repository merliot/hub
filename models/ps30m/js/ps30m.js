import { WebSocketController } from './common.js'

export function run(prefix, ws) {
	const ps30m = new Ps30m()
	ps30m.run(prefix, ws)
}

class Ps30m extends WebSocketController {

	constructor() {
		super()
	}

	open() {
		super.open()
		this.showStatus()
		this.showSystem()
		this.showController()
		this.showBattery()
		this.showLoadInfo()
		this.showSolar()
		this.showDaily()
		this.showHistorical()
	}

	showStatus() {
		if (this.state.Status === "OK") {
			if (this.overlay.innerHTML !== "Offline") {
				this.overlay.innerHTML = ""
			}
		} else {
			this.overlay.innerHTML = this.state.Status
		}
	}

	showSystem() {
		var ta = document.getElementById("system")
		ta.value = ""
		ta.value += "Software Version:            " + this.state.System.SWVersion + "\r\n"
		ta.value += "Batt Voltage Multiplier:     " + this.state.System.BattVoltMulti
	}

	showController() {
		var ta = document.getElementById("controller")
		ta.value = ""
		ta.value += "* Current (A):               " + this.state.Controller.Amps
	}

	showBattery() {
		var ta = document.getElementById("battery")
		ta.value = ""
		ta.value += "* Voltage (V):               " + this.state.Battery.Volts + "\r\n"
		ta.value += "* Current (A):               " + this.state.Battery.Amps + "\r\n"
		ta.value += "* Sense Voltage (V):         " + this.state.Battery.SenseVolts + "\r\n"
		ta.value += "* Slow Filter Voltage (V):   " + this.state.Battery.SlowVolts + "\r\n"
		ta.value += "* Slow Filter Current (A):   " + this.state.Battery.SlowAmps
	}

	showLoadInfo() {
		var ta = document.getElementById("load")
		ta.value = ""
		ta.value += "* Voltage (V):               " + this.state.LoadInfo.Volts + "\r\n"
		ta.value += "* Current (A):               " + this.state.LoadInfo.Amps
	}

	showSolar() {
		var ta = document.getElementById("solar")
		ta.value = ""
		ta.value += "* Voltage (V):               " + this.state.Solar.Volts + "\r\n"
		ta.value += "* Current (A):               " + this.state.Solar.Amps
	}

	showDaily() {
		var ta = document.getElementById("daily")
		ta.value = ""
	}

	showHistorical() {
		var ta = document.getElementById("historical")
		ta.value = ""
	}

	handle(msg) {
		switch(msg.Path) {
		case "update/status":
			this.state.Status = msg.Status
			this.showStatus()
			break
		case "update/system":
			this.state.System = msg.System
			this.showSystem()
			break
		case "update/controller":
			this.state.Controller = msg.Controller
			this.showController()
			break
		case "update/battery":
			this.state.Battery = msg.Battery
			this.showBattery()
			break
		case "update/load":
			this.state.LoadInfo = msg.LoadInfo
			this.showLoadInfo()
			break
		case "update/solar":
			this.state.Solar = msg.Solar
			this.showSolar()
			break
		case "update/daily":
			this.state.Daily = msg.Daily
			this.showDaily()
			break
		case "update/historical":
			this.state.Historical = msg.Historical
			this.showHistorical()
			break
		}
	}
}
