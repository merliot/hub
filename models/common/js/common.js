var state
var conn
var pingID

var overlay = document.getElementById("overlay")

function showDevice() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "ID:      " + state.Id + "\r\n"
	system.value += "Model:   " + state.Model + "\r\n"
	system.value += "Name:    " + state.Name
}

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

	// Create an SVG spinner element with animation
	const spinner = document.createElementNS("http://www.w3.org/2000/svg", "svg");
	spinner.setAttribute("width", "20");
	spinner.setAttribute("height", "20");
	spinner.innerHTML = '<circle cx="10" cy="10" r="7" stroke="gray" stroke-width="2" fill="transparent"></circle>';
	spinner.style.animation = "spin 1s linear infinite"; // Add animation style
	event.target.parentNode.insertBefore(spinner, event.target.nextSibling);

	fetch(downloadURL)
	.then(response => {
		// Extract the filename from Content-Disposition header
		const contentDisposition = response.headers.get('Content-Disposition')
		const match = contentDisposition.match(/filename=([^;]+)/)
		const filename = match ? match[1] : 'downloaded-file';  // Use a default filename if not found
		return Promise.all([response.blob(), filename])
	})
	.then(([blob, filename]) => {
		// Remove the spinner
		spinner.parentNode.removeChild(spinner)
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
		// Remove the spinner on error
		spinner.parentNode.removeChild(spinner)
	})
}
