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
		for (let id in this.state.Children) {
			this.loadViewTile(id)
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

	create() {
		alert("create")
	}

	showNewDialog() {
		var dialog = document.getElementById("new-dialog")
		var close = document.getElementById("new-close")
		var create = document.getElementById("new-create")

		close.onclick = () => dialog.close()
		create.onclick = () => this.create()

		dialog.showModal()
	}
}
