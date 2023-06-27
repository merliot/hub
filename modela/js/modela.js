var state
var conn
var online = false

var relays = []
for (var i = 0; i < 4; i++) {
	relays[i] = document.getElementById("relay" + i)
}

var buttons = document.getElementById("buttons")

function showRelays() {
	for (var i = 0; i < relays.length; i++) {
		relays[i].disabled = !online
	}
	buttons.style.display = "block"
}

function show() {
	overlay = document.getElementById("overlay")
	overlay.style.display = online ? "none" : "block"
	showRelays()
}

function sendClick(relay, num) {
	conn.send(JSON.stringify({Path: "click", Relay: num, State: relay.checked}))
}

function run(ws) {

	console.log('connecting...')
	conn = new WebSocket(ws)

	conn.onopen = function(evt) {
		console.log("open")
		conn.send(JSON.stringify({Path: "get/state"}))
	}

	conn.onclose = function(evt) {
		console.log("close")
		online = false
		show()
		setTimeout(run(ws), 1000)
	}

	conn.onerror = function(err) {
		console.log("error", err)
		conn.close()
	}

	conn.onmessage = function(evt) {
		msg = JSON.parse(evt.data)
		console.log('connect', msg)

		switch(msg.Path) {
		case "state":
			online = true
			state = msg
			show()
			break
		case "click":
			relays[msg.Relay].checked = msg.State
			break
		}
	}
}

