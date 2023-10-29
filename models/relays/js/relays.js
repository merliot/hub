import { WebSocketController } from './common.js'

export function run(prefix, ws) {
	const relays = new Relays()
	relays.run(prefix, ws)
}

class Relays extends WebSocketController {

	constructor() {
		super()
	}

	open() {
		super.open()
		this.showRelays()
	}

	handle(msg) {
		switch(msg.Path) {
		case "click":
			this.saveClick(msg)
			break
		}
	}

	showRelays() {
		for (let i = 0; i < 4; i++) {
			let div = document.getElementById("relay" + i)
			let label = document.getElementById("relay" + i + "-name")
			let image = document.getElementById("relay" + i + "-img")
			image.onclick = () => {
				this.relayClick(image, i)
			}
			var relay = this.state.Relays[i]
			if (relay.Gpio === "") {
				div.style.display = "none"
				label.textContent = "<unused>"
				image.src = "images/relay-off.png"
			} else {
				div.style.display = "flex"
				label.textContent = relay.Name
				this.setRelayImg(relay, image)
			}
		}
	}

	setRelayImg(relay, image) {
		if (relay.State) {
			image.src = "images/relay-on.png"
		} else {
			image.src = "images/relay-off.png"
		}
	}

	saveClick(msg) {
		var image = document.getElementById("relay" + msg.Relay + "-img")
		var relay = state.Relays[msg.Relay]
		relay.State = msg.State
		this.setRelayImg(relay, image)
	}

	relayClick(image, index) {
		var relay = this.state.Relays[index]
		relay.State = !relay.State
		this.setRelayImg(relay, image)
		this.conn.send(JSON.stringify({Path: "click", Relay: index, State: relay.State}))
	}
}
