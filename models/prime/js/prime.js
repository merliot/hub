import { WebSocketController } from './common.js'

export function run(prefix, ws) {
	const prime = new Prime()
	prime.run(prefix, ws)
}

class Prime extends WebSocketController {

	constructor() {
		super()
		this.view = document.getElementById("view")
	}

	open() {
		super.open()
		document.title = this.state.Device.Model + " - " + this.state.Device.Name
		this.view.data = "/" + this.state.Device.Id + "/"
	}

	handle(msg) {
		switch(msg.Path) {
		case "connected":
			this.connected(msg.Id)
			break
		case "disconnected":
			this.disconnected(msg.Id)
			break
		}
	}

	connected(id) {
		this.setDeviceIcon(id, true)
	}

	disconnected(id) {
		this.setDeviceIcon(id, false)
	}

	setDeviceIcon(id, online) {
	}
}
