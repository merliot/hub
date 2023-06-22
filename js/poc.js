function clearScreen() {
	const container = document.querySelector('.flex-row')
	const iframes = container.querySelectorAll('iframe')

	for (let i = 0; i < iframes.length; i++) {
		const iframe = iframes[i]
		iframe.src = ""
	}
}

function saveState(msg) {
	for (const id in msg.Children) {
		child = msg.Children[id]
		update(child)
	}
}

function update(child) {
	let iframe = document.getElementById(child.Model)
	if (iframe) {
		iframe.src = "/" + encodeURIComponent(child.Id) + "/"
	}
}

function Run(ws) {

	var conn

	function connect() {
		conn = new WebSocket(ws)

		conn.onopen = function(evt) {
			clearScreen()
			conn.send(JSON.stringify({Path: "get/state"}))
		}

		conn.onclose = function(evt) {
			clearScreen()
			setTimeout(connect, 1000)
		}

		conn.onerror = function(err) {
			conn.close()
		}

		conn.onmessage = function(evt) {
			var msg = JSON.parse(evt.data)

			console.log('demo', msg)

			switch(msg.Path) {
			case "state":
				saveState(msg)
				break
			case "connected":
			case "disconnected":
				update(msg)
				break
			}
		}
	}

	connect()
}
