var overlay = document.getElementById("overlay")

function init() {
}

function open() {
	state.Online ? online() : offline()
}

function close() {
	offline()
}

function sendClick(relay, num) {
	conn.send(JSON.stringify({Path: "click", Relay: num, State: relay.checked}))
}

function saveClick(msg) {
	relays[msg.Relay].checked = msg.State
}

function online() {
	overlay.innerHTML = ""
	for (var i = 1; i <= 4; i++) {
		checkbox = document.getElementById("relay" + i)
		label = document.querySelector('label[for="relay'+i+'"]')
		relay = state.Relays[i - 1]
		label.textContent = relay.Name
		if (relay.Name === "") {
			checkbox.style.display = "none"
			checkbox.checked = false
			checkbox.disabled = true
		} else {
			checkbox.style.display = "block"
			checkbox.checked = relay.State
			checkbox.disabled = false
		}
	}
}

function offline() {
	for (var i = 1; i <= 4; i++) {
		checkbox = document.getElementById("relay" + i)
		checkbox.disabled = true
	}
	overlay.innerHTML = "Offline"
}

function handle(msg) {
	switch(msg.Path) {
	case "click":
		saveClick(msg)
		break
	}
}
