import { WebSocketController, ViewMode } from './common.js'

export function run(prefix, url, viewMode) {
	const prime = new Hub(prefix, url, viewMode)
}

class Hub extends WebSocketController {

	constructor(prefix, url, viewMode) {
		super(prefix, url, viewMode)
		this.view = document.getElementById("view")
		this.devices = document.getElementById("devices")
		this.new = document.getElementById("new")
		this.back = document.getElementById("back")

		this.devices.onclick = () => {
			this.activeId = ""
			this.loadView()
		}

		this.new.onclick = () => {
			this.showNewDialog()
		}

		this.activeId = ''
	}

	open() {
		super.open()
		this.loadView()
	}

	handle(msg) {
		switch(msg.Path) {
		case "created/device":
			this.state.Devices[msg.Id] = {Model: msg.Model, Name: msg.Name, Online: false}
			this.activeId = msg.Id
			this.loadView()
			break
		case "deleted/device":
			//this.removeDevice(msg.Id)
			break
		}
	}

	loadView() {
		this.view.textContent = ''
		if (this.activeId === '') {
			this.loadViewTiled()
		} else {
			this.loadViewFull()
		}
	}

	loadViewFull() {
		var device = document.createElement("object")
		device.classList.add("device-full")
		device.type = "text/html"
		device.data = "/" + this.activeId + "/"

		device.onload = () => {
			device.style.height = device.contentDocument.documentElement.scrollHeight + 'px';
		}

		this.view.appendChild(device)
		this.back.classList.replace("hidden", "visible")
	}

	loadViewTiled() {
		const sortedIds = Object.keys(this.state.Children).sort((a, b) => {
			const nameA = this.state.Children[a].Name.toUpperCase();
			const nameB = this.state.Children[b].Name.toUpperCase();
			if (nameA < nameB) {
				return -1;
			}
			if (nameA > nameB) {
				return 1;
			}
			return 0;
		});
		for (let i in sortedIds) {
			this.loadViewTile(sortedIds[i])
		}
		this.back.classList.replace("visible", "hidden")
	}

	loadViewTile(id) {
		var div = document.createElement("div")
		div.classList.add("device-tile-div")

		var device = document.createElement("object")
		device.classList.add("device-tile")
		device.type = "text/html"
		device.data = "/" + id + "/tile"

		div.onclick = () => {
			this.activeId = id;
			this.loadView();
		}

		div.appendChild(device)
		this.view.appendChild(div)
	}

	generateRandomId() {
		// Generate 4 bytes of random data for ID; return as hex-encoded string
		return 'xxxxxxxx'.replace(/[x]/g, function(c) {
			const r = Math.floor(Math.random() * 16);
			return r.toString(16);
		});
	}

	create() {
		let myuuid = this.generateRandomId()
		console.log('Your UUID is: ' + myuuid);
	}

	loadModels() {
		var xhr = new XMLHttpRequest();
		xhr.onreadystatechange = function () {
			if (xhr.readyState == 4 && xhr.status == 200) {
				document.getElementById("new-dialog-models").innerHTML = xhr.responseText;
			}
		};
		xhr.open("GET", "/models", true);
		xhr.send();
	}

	showNewDialog() {
		var dialog = document.getElementById("new-dialog")
		var close = document.getElementById("new-close")
		var create = document.getElementById("new-create")

		close.onclick = () => dialog.close()
		create.onclick = () => this.create()

		this.loadModels()
		dialog.showModal()
	}
}
