let chart = new Chart(document.getElementById("chart"), {
	data: {
		labels: [],
		datasets: [{
			type: 'line',
			label: 'Array Current',
			data: [],
			pointStyle: false,
		}, {
			type: 'bar',
			label: 'Battery Terminal Voltage',
			data: [],
		}, {
			type: 'line',
			label: 'Line Current',
			data: [],
			pointStyle: false,
		}],
	},
	options: {
		animation: false,
		responsive: true,
		maintainAspectRatio: false,
		scales: {
			x: {
				reverse: true,
			},
		},
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

function showChartRecords(array, size) {
	chart.data.labels = Array(size).fill("");
	for (let j = 0; j < chart.data.datasets.length; j++) {
		chart.data.datasets[j].data = Array(size).fill(null);
		for (let i = 0; i < array.length; i++) {
			chart.data.datasets[j].data[i] = array[i][j].toFixed(2)
		}
	}
	chart.update()
}

function showChart() {
	let btns = document.getElementsByName("period")
	for (i = 0; i < btns.length; i++) {
		if (btns[i].checked) {
			switch (btns[i].value) {
			case "minute":
				showChartRecords(state.Seconds, 60)
				break
			case "hour":
				showChartRecords(state.Minutes, 60)
				break
			case "day":
				showChartRecords(state.Hours, 24)
				break
			case "year":
				showChartRecords(state.Days, 365)
				break
			}
			return
		}
	}
}

function show() {
	showSystem()
	showChart()
}

function hide() {
}

function saveRecord(array, record, size) {
	array.unshift(record)
	if (array.length > size) {
		array.pop()
	}
}

function handle(msg) {
	switch(msg.Path) {
			/*
	case "regs/update":
		showRegs(msg)
		break
		*/
	case "update/second":
		saveRecord(state.Seconds, msg.Record, 60)
		showChart()
		break
	case "update/minute":
		saveRecord(state.Minutes, msg.Record, 60)
		showChart()
		break
	case "update/hour":
		saveRecord(state.Hours, msg.Record, 24)
		showChart()
		break
	case "update/day":
		saveRecord(state.Days, msg.Record, 365)
		showChart()
		break
	}
}
