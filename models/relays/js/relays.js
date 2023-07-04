var relays = []

function init() {
	for (var i = 0; i < 4; i++) {
		relays[i] = document.getElementById("relay" + i)
	}
}

function show() {
	for (var i = 0; i < relays.length; i++) {
		relays[i].checked = state.States[i]
		relays[i].disabled = false
	}
}

function hide() {
	for (var i = 0; i < relays.length; i++) {
		relays[i].disabled = true
	}
}

function sendClick(relay, num) {
	conn.send(JSON.stringify({Path: "click", Relay: num, State: relay.checked}))
}

function saveClick(msg) {
	relays[msg.Relay].checked = msg.State
}

function handle(msg) {
	switch(msg.Path) {
	case "click":
		saveClick(msg)
		break
	}
}
