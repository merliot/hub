class WebSocketController {

	constructor() {
		this.state = null;
		this.conn = null;
		this.pingID = null;
		this.pingAlive = false;
		this.pingSent = null;
		this.overlay = document.getElementById("overlay");
	}

	run(prefix, ws) {

		const url = new URL(ws);
		const params = new URLSearchParams(url.search);
		const pingPeriod = params.get("ping-period") * 1000;

		console.log(prefix, 'connecting...');
		this.conn = new WebSocket(ws);

		this.conn.onopen = (evt) => {
			console.log(prefix, 'open');
			this.conn.send(JSON.stringify({Path: "get/state"}));
			this.pingAlive = true;
			this.pingID = setInterval(() => this.ping(prefix), pingPeriod)
		};

		this.conn.onclose = (evt) => {
			console.log(prefix, 'close');
			this.close();
			clearInterval(this.pingID);
			setTimeout(() => this.run(prefix, ws), 1000); // Reconnecting after 1 second
		};

		this.conn.onerror = (err) => {
			console.log(prefix, 'error', err);
			this.conn.close();
		};

		this.conn.onmessage = (evt) => {

			if (evt.data == "pong") {
				//console.log(prefix, "pong", new Date() - this.pingSent)
				this.pingAlive = true
				return
			}

			var msg = JSON.parse(evt.data)
			console.log(prefix, msg)

			switch(msg.Path) {
				case "state":
					this.state = msg
					this.open()
					break
				case "online":
					this.state.Online = true
					this.online()
					break
				case "offline":
					this.state.Online = false
					this.offline()
					break
				default:
					this.handle(msg)
					break
			}
		};
	}

	ping(prefix) {
		if (!this.pingAlive) {
			console.log(prefix, "NOT ALIVE", new Date() - this.pingSent)
			// This waits for an ACK from server, but the server
			// may be gone, it may take a bit to close the websocket
			this.conn.close()
			return
		}
		this.pingAlive = false
		this.conn.send("ping")
		this.pingSent = new Date()
	}

	open() {
		this.state.Online ? this.online() : this.offline()
	}

	close() {
		this.offline()
	}

	online() {
		this.overlay.innerHTML = ""
	}

	offline() {
		this.overlay.innerHTML = "Offline"
	}

	handle(msg) {
		// drop msg
	}
}

export { WebSocketController };

function downloadFile(event) {
	event.preventDefault()
	var downloadURL = event.target.innerText

	var response = document.getElementById("download-response")
	response.innerText = ""

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

			// Extract Content-MD5 header and decode from base64
			const base64Md5 = response.headers.get("Content-MD5")
			const md5sum = atob(base64Md5)

			// Extract the filename from Content-Disposition header
			const match = contentDisposition.match(/filename=([^;]+)/)
			const filename = match ? match[1] : 'downloaded-file';  // Use a default filename if not found
			return Promise.all([response.blob(), filename, md5sum])
		})
		.then(([blob, filename, md5sum]) => {
			// Create a temporary link element to trigger the download
			const a = document.createElement('a')
			a.href = URL.createObjectURL(blob)
			a.style.display = 'none'
			a.download = filename
			document.body.appendChild(a)
			a.click();  // Simulate a click on the link
			document.body.removeChild(a)
			gopher.style.display = "none"
			response.innerText = "MD5: " + md5sum
			response.style.color = "black"
		})
		.catch(error => {
			console.error('Error downloading file:', error)
			gopher.style.display = "none"
			response.innerText = error
			response.style.color = "red"
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
	const params = new URLSearchParams(deployParams)

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

function updateInstructions(target) {
	var instructions = document.getElementById('deploy-instructions')
	var xhr = new XMLHttpRequest();
	xhr.open('GET', "docs/install/" + target + ".md", true);
	xhr.onreadystatechange = function() {
		if (this.readyState == 4 && this.status == 200) {
			instructions.innerHTML = this.responseText;
		} else {
			instructions.innerHTML = ""
		}
	};
	xhr.send();
}

function updateLocalHttpServer(target) {
	var div = document.getElementById('deploy-http-div')
	var http = document.getElementById('deploy-http')
	switch (target) {
		case "demo":
		case "x86-64":
		case "rpi":
			div.style.display = "flex"
			http.disabled = false
			break
		default:
			div.style.display = "none"
			http.disabled = true
			http.checked = false
			break
	}
}

function updateSsid(target) {
	var div = document.getElementById('deploy-ssid-div')
	var ssid = document.getElementById('deploy-ssid')
	switch (target) {
		case "demo":
		case "x86-64":
		case "rpi":
			div.style.display = "none"
			ssid.disabled = true
			ssid.name = ""
			break
		default:
			div.style.display = "flex"
			ssid.disabled = false
			ssid.name = "ssid"
			break
	}
}

function handleBackup(backup, first) {
	var backupHub = document.getElementById("deploy-backuphub")
	if (first) {
		if (backupHub.value !== "") {
			backup.checked = true
		}
	}
	if (backup.checked) {
		backupHub.disabled = false;
		backupHub.name = "backuphub";
	} else {
		backupHub.disabled = true;
		backupHub.name = "";
	}
	updateDeployLink()
}

function handleTarget(target) {
	updateInstructions(target)
	updateLocalHttpServer(target)
	updateSsid(target)
	updateDeployLink()
}

function stageDeploy(deployParams) {

	stageFormData(deployParams)

	document.getElementById("download-link").addEventListener("click", downloadFile)

	var backup = document.getElementById("deploy-backup")
	backup.addEventListener("change", function() { handleBackup(backup, false) })
	handleBackup(backup, true)

	// Attach an event listener to the deploy-target dropdown to set instructions
	var target = document.getElementById('deploy-target')
	target.addEventListener('change', function() { handleTarget(this.value) })
	handleTarget(target.value)

	var form = document.getElementById("deploy-form")
	form.addEventListener('input', function (event) { updateDeployLink() })

	updateDeployLink()
}

export { stageDeploy };
