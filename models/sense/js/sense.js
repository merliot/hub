var state
var conn
var online = false

function showSystem() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "ID:      " + state.Identity.Id + "\r\n"
	system.value += "Model:   " + state.Identity.Model + "\r\n"
	system.value += "Name:    " + state.Identity.Name
}

function show() {
	overlay = document.getElementById("overlay")
	overlay.style.display = online ? "none" : "block"
	showSystem()
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
		}
	}
}

