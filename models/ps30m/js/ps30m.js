function init() {
}

function chargeState(state) {
	const names = ["START", "NIGHT_CHECK", "NIGHT_CHECK", "NIGHT",
		"FAULT", "BULK", "ABSORPTION", "FLOAT", "EQUALIZE"]
	return names[state]
}

function loadState(state) {
	const names = ["START", "LOAD_ON", "LVD_WARNING", "LVD",
		"FAULT", "DISCONNECT", "LOAD_OFF", "OVERRIDE"]
	return names[state]
}

function showRegs(msg) {
	let regs = document.getElementById("regs")
	regs.value = ""
	regs.value += "Array Current (A):            " + msg.Regs[17].toFixed(2) + "\r\n"
	regs.value += "Battery Terminal Voltage (V): " + msg.Regs[18].toFixed(2) + "\r\n"
	regs.value += "Load Current (A):             " + msg.Regs[22].toFixed(2) + "\r\n"
	regs.value += "Charge State:                 " + chargeState(msg.Regs[33]) + "\r\n"
	regs.value += "Load State:                   " + loadState(msg.Regs[46]) + "\r\n"
}

function show() {
	showSystem()
}

function hide() {
}

function handle(msg) {
	switch(msg.Path) {
	case "regs/update":
		showRegs(msg)
		break
	}
}
