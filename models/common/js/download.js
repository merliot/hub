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

function deployLink() {
	var form = document.getElementById("deploy-form")

	var currentURL = window.location.href
	var lastIndex = currentURL.lastIndexOf('/');
	var baseURL = currentURL.substring(0, lastIndex);

	var formData = new FormData(form)
	var query = new URLSearchParams(formData).toString()
	var linkURL = "/deploy?" + query

	return baseURL + linkURL
}

function downloadFile(event) {
	event.preventDefault()

	var response = document.getElementById("download-response")
	response.innerText = ""

	var gopher = document.getElementById("gopher")
	gopher.style.display = "block"

	fetch(deployLink())
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
}

function handleHttp(http, first) {
	var port = document.getElementById("deploy-port")
	if (first) {
		if (port.value !== "") {
			http.checked = true
		}
	}
	if (http.checked) {
		port.disabled = false;
		port.name = "port";
	} else {
		port.disabled = true;
		port.name = "";
	}
}

function updateHttp(target) {
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

function handleTarget(target) {
	updateHttp(target)
	updateSsid(target)
}

function stageDeploy(deployParams) {

	stageFormData(deployParams)

	document.getElementById("download-btn").addEventListener("click", downloadFile)

	var backup = document.getElementById("deploy-backup")
	backup.addEventListener("change", function() { handleBackup(backup, false) })
	handleBackup(backup, true)

	var http = document.getElementById("deploy-http")
	http.addEventListener("change", function() { handleHttp(http, false) })
	handleHttp(http, true)

	// Attach an event listener to the deploy-target dropdown
	var target = document.getElementById('deploy-target')
	target.addEventListener('change', function() { handleTarget(this.value) })
	handleTarget(target.value)
}

export { stageDeploy };

