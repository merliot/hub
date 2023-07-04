var state
var conn
var pingID

var overlay = document.getElementById("overlay")

function showSystem() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "ID:      " + state.Identity.Id + "\r\n"
	system.value += "Model:   " + state.Identity.Model + "\r\n"
	system.value += "Name:    " + state.Identity.Name
}

function offline() {
	overlay.style.display = "block"
	hide()
	clearInterval(pingID)
}

function ping() {
	conn.send("ping")
}

function online() {
	showSystem()
	show()
	overlay.style.display = "none"
	// for Koyeb work-around, ping every 60s to keep websocket alive
	pingID = setInterval(ping, 1 * 60 * 1000)
}

function run(prefix, ws) {

	init()

	console.log(prefix, 'connecting...')
	conn = new WebSocket(ws)

	conn.onopen = function(evt) {
		console.log(prefix, 'open')
		conn.send(JSON.stringify({Path: "get/state"}))
	}

	conn.onclose = function(evt) {
		console.log(prefix, 'close')
		offline()
		setTimeout(run(prefix, ws), 1000)
	}

	conn.onerror = function(err) {
		console.log(prefix, 'error', err)
		conn.close()
	}

	conn.onmessage = function(evt) {
		msg = JSON.parse(evt.data)
		console.log(prefix, msg)

		switch(msg.Path) {
		case "state":
			state = msg
			online()
			break
		default:
			handle(msg)
			break
		}
	}
}
