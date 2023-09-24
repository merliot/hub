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

	var downloadError = document.getElementById("download-error")
	downloadError.style.display = "none"

	var gopher = document.getElementById("gopher")
	gopher.style.display = "block"

	fetch(downloadURL)
	.then(response => {
		if (!response.ok) {
			// If we didn't get a 2xx response, throw an error with the response text
			return response.text().then(text => { throw new Error(text) })
		}

		const contentDisposition = response.headers.get('Content-Disposition')
		if (!contentDisposition) {
			throw new Error('Content-Disposition header missing')
		}

		// Extract the filename from Content-Disposition header
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
		gopher.style.display = "none"
	})
	.catch(error => {
		console.error('Error downloading file:', error)
		downloadError.innerText = error
		downloadError.style.display = "block"
		gopher.style.display = "none"
	})
}

function updateDeployLink() {
	var link = document.getElementById("download-link")
	var form = document.getElementById("deploy-form")

	var currentURL = window.location.href
	var lastIndex = currentURL.lastIndexOf('/');
	var baseURL = currentURL.substring(0, lastIndex);

	var formData = new FormData(form)
	var query = new URLSearchParams(formData).toString()
	var linkURL = "/deploy?" + query

	var downloadURL = baseURL + linkURL
	link.innerHTML = downloadURL
}

function decodeHtmlEntities(input) {
	var doc = new DOMParser().parseFromString(input, 'text/html');
	return doc.documentElement.textContent;
}

function stageFormData(deployParams) {
	var form = document.getElementById("deploy-form")
	const params = new URLSearchParams(decodeHtmlEntities(deployParams))

	params.forEach((value, key) => {
		let element = form.elements[key];
		if (element) {
			switch (element.type) {
			case 'checkbox':
				element.checked = value === 'on';
				break;
			case 'radio':
				// If there are multiple radio buttons with the
				// same name, value will determine which one to check
				element = [...form.elements[key]].find(radio => radio.value === value);
				if (element) element.checked = true;
				break;
			default:
				element.value = value;
				break;
			}
			// Manually dispatch a change event
			let event = new Event('change', {});
			element.dispatchEvent(event);
		}
	});
}

function stageDeploy(deployParams) {

	stageFormData(deployParams)

	document.getElementById("download-link").addEventListener("click", downloadFile)

	var backup = document.getElementById("deploy-backup")
	var backupHub = document.getElementById("deploy-backup-hub")

	backup.addEventListener("change", function() {
		if (this.checked) {
			backupHub.disabled = false;
			backupHub.name = "backup-hub";
		} else {
			backupHub.disabled = true;
			backupHub.name = "";
		}
		updateDeployLink()
	})

	var form = document.getElementById("deploy-form")
	form.addEventListener('input', function (event) {
		updateDeployLink()
	})
	updateDeployLink()
}
