let acChart = new Chart(document.getElementById("array-current"), {
	data: {
		labels: [],
		datasets: [{
			type: 'bar',
			label: 'Array Current',
			data: [],
		}],
	},
	options: {
		animation: false,
	},
});

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

function showArrayCurrent() {
	acChart.data.labels = Array(60).fill("");
	acChart.data.datasets[0].data = Array(60).fill(null);
	acChart.update()
}

function show() {
	showSystem()
	showArrayCurrent()
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
