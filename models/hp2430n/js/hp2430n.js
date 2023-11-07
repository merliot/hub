import { WebSocketController } from './common.js'

export function run(prefix, ws) {
	const hp2430n = new Hp2430n()
	hp2430n.run(prefix, ws)
}

class Hp2430n extends WebSocketController {

	constructor() {
		super()
		this.chart = new Chart(document.getElementById("chart"), {
			data: {
				labels: [],
				datasets: [{
					type: 'line',
					label: 'Battery Voltage',
					data: [],
					pointStyle: false,
				}, {
					type: 'line',
					label: 'Charging Current',
					data: [],
					pointStyle: false,
				}, {
					type: 'line',
					label: 'Load Voltage',
					data: [],
					pointStyle: false,
				}, {
					type: 'line',
					label: 'Load Current',
					data: [],
					pointStyle: false,
				}, {
					type: 'line',
					label: 'Solar Voltage',
					data: [],
					pointStyle: false,
				}, {
					type: 'line',
					label: 'Solar Current',
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
		var btns = document.getElementsByName("period")
		for (var i = 0; i < btns.length; i++) {
			btns[i].onclick = () => this.showChart()
		}
	}

	open() {
		super.open()
		this.showStatus()
		this.showChart()
	}

	handle(msg) {
		switch(msg.Path) {
		case "update/status":
			this.state.ChargeState = msg.ChargeState
			this.state.LoadState = msg.LoadState
			this.showStatus()
			break
		case "update/second":
			this.saveRecord(this.state.Seconds, msg.Record, 60)
			this.showChart()
			break
		case "update/minute":
			this.saveRecord(this.state.Minutes, msg.Record, 60)
			this.showChart()
			break
		case "update/hour":
			this.saveRecord(this.state.Hours, msg.Record, 24)
			this.showChart()
			break
		case "update/day":
			this.saveRecord(this.state.Days, msg.Record, 365)
			this.showChart()
			break
		}
	}

	saveRecord(array, record, size) {
		array.unshift(record)
		if (array.length > size) {
			array.pop()
		}
	}

	chargeState(state) {
		const names = ["START", "NIGHT_CHECK", "NIGHT_CHECK", "NIGHT",
			"FAULT", "BULK", "ABSORPTION", "FLOAT", "EQUALIZE"]
		return names[state]
	}

	loadState(state) {
		const names = ["START", "LOAD_ON", "LVD_WARNING", "LVD",
			"FAULT", "DISCONNECT", "LOAD_OFF", "OVERRIDE"]
		return names[state]
	}

	showStatus() {
		var textarea = document.getElementById("status")
		textarea.value = ""
		textarea.value += "Charge State:   " + this.chargeState(this.state.ChargeState) + "\r\n"
		textarea.value += "Load State:     " + this.loadState(this.state.LoadState) + "\r\n"
	}

	showChartRecords(array, size) {
		this.chart.data.labels = Array(size).fill("");
		for (let j = 0; j < this.chart.data.datasets.length; j++) {
			this.chart.data.datasets[j].data = Array(size).fill(null);
			for (let i = 0; i < array.length; i++) {
				this.chart.data.datasets[j].data[i] = array[i][j].toFixed(2)
			}
		}
		this.chart.update()
	}

	showChart() {
		var btns = document.getElementsByName("period")
		for (var i = 0; i < btns.length; i++) {
			if (btns[i].checked) {
				switch (btns[i].value) {
				case "minute":
					this.showChartRecords(this.state.Seconds, 60)
					break
				case "hour":
					this.showChartRecords(this.state.Minutes, 60)
					break
				case "day":
					this.showChartRecords(this.state.Hours, 24)
					break
				case "year":
					this.showChartRecords(this.state.Days, 365)
					break
				}
				return
			}
		}
	}
}
