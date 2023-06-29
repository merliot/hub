var state
var conn
var online = false

var relays = []
for (var i = 0; i < 4; i++) {
	relays[i] = document.getElementById("relay" + i)
}

var buttons = document.getElementById("buttons")

function showSystem() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "ID:      " + state.Identity.Id + "\r\n"
	system.value += "Model:   " + state.Identity.Model + "\r\n"
	system.value += "Name:    " + state.Identity.Name
}

function showRelays() {
	for (var i = 0; i < relays.length; i++) {
		relays[i].disabled = !online
	}
	buttons.style.display = "block"
}

function show() {
	overlay = document.getElementById("overlay")
	overlay.style.display = online ? "none" : "block"
	showSystem()
	showRelays()
}

function saveState(msg) {
	state = msg
	for (var i = 0; i < relays.length; i++) {
		relays[i].checked = msg.States[i]
	}
}

function sendClick(relay, num) {
	conn.send(JSON.stringify({Path: "click", Relay: num, State: relay.checked}))
}

function saveClick(msg) {
	relays[msg.Relay].checked = msg.State
}

function run(ws) {

	console.log('[relays]', 'connecting...')
	conn = new WebSocket(ws)

	conn.onopen = function(evt) {
		console.log('[relays]', 'open')
		conn.send(JSON.stringify({Path: "get/state"}))
	}

	conn.onclose = function(evt) {
		console.log('[relays]', 'close')
		online = false
		show()
		setTimeout(run(ws), 1000)
	}

	conn.onerror = function(err) {
		console.log('[relays]', 'error', err)
		conn.close()
	}

	conn.onmessage = function(evt) {
		msg = JSON.parse(evt.data)
		console.log('[relays]', msg)

		switch(msg.Path) {
		case "state":
			online = true
			saveState(msg)
			show()
			break
		case "click":
			saveClick(msg)
			break
		}
	}
}

