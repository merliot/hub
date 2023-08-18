var state
var conn
var pingID

var overlay = document.getElementById("overlay")

function wsclose() {
	close()
	clearInterval(pingID)
}

function ping() {
	conn.send("ping")
}

function wsopen() {
	// for Koyeb work-around, ping every 60s to keep websocket alive
	pingID = setInterval(ping, 1 * 60 * 1000)
	open()
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
		wsclose()
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
			wsopen()
			break
		case "online":
			state.Online = true
			online()
			break
		case "offline":
			state.Online = false
			offline()
			break
		default:
			handle(msg)
			break
		}
	}
}

function downloadFile(event) {
	event.preventDefault()
	var downloadURL = event.target.innerText

	// TODO spinner animation when downloading file

	fetch(downloadURL)
	.then(response => {
		// Extract the filename from Content-Disposition header
		const contentDisposition = response.headers.get('Content-Disposition')
		const match = contentDisposition.match(/filename=([^;]+)/)
		const filename = match ? match[1] : 'downloaded-file';  // Use a default filename if not found
		return Promise.all([response.blob(), filename])
	})
	.then(([blob, filename]) => {
		// Create a temporary link element to trigger the download
		const a = document.createElement('a')
		a.href = URL.createObjectURL(blob)
		a.style.display = 'none'
		a.download = filename
		document.body.appendChild(a)
		a.click();  // Simulate a click on the link
		document.body.removeChild(a)
	})
	.catch(error => {
		console.error('Error downloading file:', error)
	})
}
