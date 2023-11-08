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

	loadStatus(loadInfo) {
		var load = (loadInfo >> 8) & 0x80
		if (load == 0) {
			return "OFF"
		} else {
			return "ON"
		}
	}

	loadBrightness(loadInfo) {
		var bright = (loadInfo >> 8) & 0x7F
		var percent = (bright / 0x64) * 100
		return `${percent.toFixed(0)}%`
	}

	batteryStatus(loadInfo) {
		switch(loadInfo & 0xff) {
		case 0:
			return "Charging Deactivated"
			break
		case 1:
			return "Charging Activated"
			break
		case 2:
			return "MPPT Charging Mode"
			break
		case 3:
			return "Equalizing Charging Mode"
			break
		case 4:
			return "Boost Charging Mode"
			break
		case 5:
			return "Floating Charging Mode"
			break
		case 6:
			return "Constant Current (overpower)"
			break
		}
	}

	showStatus() {
		var textarea = document.getElementById("status")
		textarea.value = ""
		textarea.value += "Load Status:     " + this.loadStatus(this.state.LoadInfo) + "\r\n"
		textarea.value += "Load Brightness: " + this.loadBrightness(this.state.LoadInfo) + "\r\n"
		textarea.value += "Battery Status:  " + this.batteryStatus(this.state.LoadInfo) + "\r\n"
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
