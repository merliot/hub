var overlay = document.getElementById("overlay")

function init() {
}

function showRelays() {
	for (var i = 0; i < 4; i++) {
		div = document.getElementById("relay" + i)
		label = document.getElementById("relay" + i + "-name")
		image = document.getElementById("relay" + i + "-img")
		relay = state.Relays[i]
		if (relay.Gpio === "") {
			div.style.display = "none"
			label.textContent = "<unused>"
			image.src = "images/relay-off.png"
		} else {
			div.style.display = "flex"
			label.textContent = relay.Name
			setRelayImg(relay, image)
		}
	}
}

function open() {
	state.Online ? online() : offline()
	showRelays()
}

function close() {
	offline()
}

function setRelayImg(relay, image) {
	if (relay.State) {
		image.src = "images/relay-on.png"
	} else {
		image.src = "images/relay-off.png"
	}
}

function saveClick(msg) {
	var image = document.getElementById("relay" + msg.Relay + "-img")
	var relay = state.Relays[msg.Relay]
	relay.State = msg.State
	setRelayImg(relay, image)
}

function relayClick(image, index) {
	var relay = state.Relays[index]
	relay.State = !relay.State
	setRelayImg(relay, image)
	conn.send(JSON.stringify({Path: "click", Relay: index, State: relay.State}))
}

function online() {
	overlay.innerHTML = ""
}

function offline() {
	overlay.innerHTML = "Offline"
}

function handle(msg) {
	switch(msg.Path) {
	case "click":
		saveClick(msg)
		break
	}
}
