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
			let gpio = document.getElementById("gpio" + i)
			var relay = this.state.Relays[i]
			if (relay.Gpio === "") {
				div.classList.replace("relay", "relay-unused")
				gpio.classList.replace("gpio", "gpio-unused")
				label.classList.replace("relay-name", "relay-name-unused")
				gpio.textContent = ""
				label.textContent = "unasigned"
				image.src = "images/relay-unused.png"
			} else {
				div.classList.replace("relay-unused", "relay")
				gpio.classList.replace("gpio-unused", "gpio")
				label.classList.replace("relay-name-unused", "relay-name")
				gpio.textContent = relay.Gpio
				label.textContent = relay.Name
				this.setRelayImg(relay, image)
				div.onclick = () => {
					this.relayClick(image, i)
				}
			}
		}
	}

	setRelayImg(relay, image) {
		image.disabled = false
		if (relay.State) {
			image.src = "images/relay-on.png"
		} else {
			image.src = "images/relay-off.png"
		}
	}

	saveClick(msg) {
		var image = document.getElementById("relay" + msg.Relay + "-img")
		var relay = this.state.Relays[msg.Relay]
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
