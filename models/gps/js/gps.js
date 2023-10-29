import { WebSocketController } from './common.js'

export function run(prefix, ws) {
	const gps = new Gps()
	gps.run(prefix, ws)
}

class Gps extends WebSocketController {

	constructor() {
		super()

		// Create a Leaflet map using OpenStreetMap
		this.map = L.map('map').setZoom(13)
		L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
			maxZoom: 19,
			attribution: '© OpenStreetMap',
		}).addTo(this.map)
		this.marker = L.marker([0, 0]).addTo(this.map)
	}

	open() {
		super.open()
		this.showMarker()
	}

	close() {
		this.map.remove()
		super.close()
	}

	handle(msg) {
		switch(msg.Path) {
		case "update":
			this.state.Lat = msg.Lat
			this.state.Long = msg.Long
			this.showMarker()
			break
		}
	}

	showMarker() {
		this.marker.setLatLng([this.state.Lat, this.state.Long])
		this.map.panTo([this.state.Lat, this.state.Long])
	}
}
