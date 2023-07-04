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
	clearInterval(pingID)
}

function ping() {
	conn.send("ping")
}

function online() {
	showSystem()
	overlay.style.display = "none"
	// for Koyeb work-around
	pingID = setInterval(ping, 1500)
}

function run(ws) {

	console.log('[led]', 'connecting...')
	conn = new WebSocket(ws)

	conn.onopen = function(evt) {
		console.log('[led]', 'open')
		conn.send(JSON.stringify({Path: "get/state"}))
	}

	conn.onclose = function(evt) {
		console.log('[led]', 'close')
		offline()
		setTimeout(run(ws), 1000)
	}

	conn.onerror = function(err) {
		console.log('[led]', 'error', err)
		conn.close()
	}

	conn.onmessage = function(evt) {
		msg = JSON.parse(evt.data)
		console.log('[led]', msg)

		switch(msg.Path) {
		case "state":
			state = msg
			online()
			break
		}
	}
}
