import { WebSocketController } from './common.js'

export function run(prefix, ws) {
	const uv = new UV()
	uv.run(prefix, ws)
}

class UV extends WebSocketController {

	constructor() {
		super()
		this.gauge = new RadialGauge({
			renderTo: document.getElementById("gauge"),
			majorTicks: [0,20,40,60,80,100,120],
			minorTicks: 10,
			highlights: [
				{from: 0, to: 24.8, color: "green"},
				{from: 24.8, to: 49.8, color: "yellow"},
				{from: 49.8, to: 66.4, color: "orange"},
				{from: 66.4, to: 91.288, color: "red"},
				{from: 91.288, to: 120, color: "violet"},
			],
			maxValue: 120,
			units: "W/(m*m)",
			title: "UV Light Intensity",
			width: 300,
			height: 300,
			valueInt: 0,
			valueDec: 3,
		})
	}

	open() {
		super.open()
		this.gauge.draw()
		this.update()
	}

	handle(msg) {
		switch(msg.Path) {
		case "update":
			this.state.Intensity = msg.Intensity
			this.state.RiskLevel = msg.RiskLevel
			this.update()
			break
		}
	}

	update() {
		this.gauge.value = this.state.Intensity / 1000.0
	}
}
