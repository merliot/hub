import { WebSocketController } from './common.js'

export function run(prefix, ws) {
	const hp2430n = new Hp2430n()
	hp2430n.run(prefix, ws)
}

class Hp2430n extends WebSocketController {

	constructor() {
		super()
	}

	open() {
		super.open()
		this.showSystem()
		this.showBattery()
		this.showLoad()
		this.showSolar()
	}

	showSystem() {
		var ta = document.getElementById("system")
		ta.value = ""
		ta.value += "Maximum Voltage Supported (V):  " + this.state.System.MaxVolts + "\r\n"
		ta.value += "Rated Charging Current (A):     " + this.state.System.ChargeAmps + "\r\n"
		ta.value += "Rated Discharging Current (A):  " + this.state.System.DischargeAmps + "\r\n"
		ta.value += "Product Type:                   " + this.state.System.ProductType + "\r\n"
		ta.value += "Model:                          " + this.state.System.Model + "\r\n"
		ta.value += "Software Version:               " + this.state.System.SWVersion + "\r\n"
		ta.value += "Hardware Version:               " + this.state.System.HWVersion + "\r\n"
		ta.value += "Serial:                         " + this.state.System.Serial + "\r\n"
		ta.value += "Temp (C):                       " + this.state.System.Temp + "\r\n"
	}

	showBattery() {
		var ta = document.getElementById("battery")
		ta.value = ""
		ta.value += "Capacity SOC:    " + this.state.Battery.SOC + "\r\n"
		ta.value += "Voltage (V):     " + this.state.Battery.Volts + "\r\n"
		ta.value += "Current (A):     " + this.state.Battery.Amps + "\r\n"
		ta.value += "Temp (C):        " + this.state.Battery.Temp + "\r\n"
		ta.value += "Charging State:  " + this.state.Battery.ChargeState + "\r\n"
	}

	showLoad() {
		var ta = document.getElementById("load")
		ta.value = ""
		ta.value += "Voltage (V):   " + this.state.LoadInfo.Volts + "\r\n"
		ta.value += "Current (A):   " + this.state.LoadInfo.Amps + "\r\n"
		ta.value += "Status:        " + this.state.LoadInfo.Status + "\r\n"
		ta.value += "Brightness:    " + this.state.LoadInfo.Brightness + "\r\n"
	}

	showSolar() {
		var ta = document.getElementById("solar")
		ta.value = ""
		ta.value += "Voltage (V):   " + this.state.Solar.Volts + "\r\n"
		ta.value += "Current (A):   " + this.state.Solar.Amps + "\r\n"
	}
}
