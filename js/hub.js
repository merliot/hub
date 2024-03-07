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
		this.menuIcon = document.getElementById("menu-icon")
		this.menu = document.getElementById("menu");
		this.menuItems = document.querySelectorAll(".menu-item");
		this.backIcon = document.getElementById("back-icon")
		this.trashIcon = document.getElementById("trash-icon")
		this.newDialog = document.getElementById("new-dialog")
		this.deleteDialog = document.getElementById("delete-dialog")

		this.devices.onclick = () => {
			this.activeId = ""
			this.loadView()
		}

		this.trashIcon.onclick = () => {
			this.showDeleteDialog()
		}

		this.new.onclick = () => {
			this.showNewDialog()
		}

		this.menuIcon.onclick = () => {
			this.toggleMenu()
		}

		this.menuItems.forEach(item => {
			item.addEventListener("click", () => {
				this.menuItemClick(item);
			});
		});

		this.activeId = ''
		this.localEvent = false
	}

	open() {
		super.open()
		this.loadView()
	}

	handle(msg) {
		switch(msg.Path) {
		case "created/device":
			this.state.Children[msg.Id] = {Model: msg.Model, Name: msg.Name, Online: false}
			if (this.localEvent) {
				this.activeId = msg.Id
				this.localEvent = false
				this.loadView()
			} else if (this.activeId === "") {
				this.loadView()
			}
			break
		case "deleted/device":
			delete this.state.Children[msg.Id]
			if (this.localEvent) {
				this.activeId = ""
				this.localEvent = false
				this.loadView()
			} else if (this.activeId === "") {
				this.loadView()
			}
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
		this.backIcon.classList.replace("hidden", "visible")
		this.trashIcon.classList.replace("hidden", "visible")
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
		this.backIcon.classList.replace("visible", "hidden")
		this.trashIcon.classList.replace("visible", "hidden")
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

	closeNewDialog(event) {
		event.preventDefault()
		this.newDialog.close()
	}

	async create(event) {
		event.preventDefault()

		var createErr = document.getElementById("create-err")
		var form = document.getElementById("new-dialog-form")
		var formData = new FormData(form)
		var query = new URLSearchParams(formData).toString()

		this.localEvent = true
		let response = await fetch("/create?" + query)

		if (response.status == 201) {
			this.newDialog.close()
		} else {
			let data = await response.text()
			createErr.innerText = data
		}
	}

	async deletef() {
		this.localEvent = true

		let response = await fetch("/delete?id=" + this.activeId)
		var deleteErr = document.getElementById("delete-err")

		if (response.status == 200) {
			this.deleteDialog.close()
		} else {
			let data = await response.text()
			deleteErr.innerText = data
		}
	}

	setupModelClick() {
		var models = document.querySelectorAll('.model');
		models.forEach(function(model) {
			model.onclick = () => {
				// Reset all models to their default style
				var models = document.querySelectorAll('.model');
				models.forEach(function(div) {
					div.classList.remove('selected');
				});
				// Mark the selected model
				model.classList.add('selected');
				// Set the selected model's html id in the hidden input
				document.getElementById('new-model').value = model.id;
			}
		});
	}

	loadModels() {
		var setupModelClick = this.setupModelClick
		var newDialog = this.newDialog
		var xhr = new XMLHttpRequest();
		xhr.onreadystatechange = function () {
			if (xhr.readyState == 4 && xhr.status == 200) {
				document.getElementById("new-dialog-models").innerHTML = xhr.responseText;
				setupModelClick()
				newDialog.showModal()
			}
		};
		xhr.open("GET", "/models", true);
		xhr.send();
	}

	generateRandomId() {
		// Generate 8 bytes of random data for ID; return as hex-encoded string
		return 'xxxxxxxx-xxxxxxxx'.replace(/[x]/g, function(c) {
			const r = Math.floor(Math.random() * 16);
			return r.toString(16);
		});
	}

	showNewDialog() {
		var close = document.getElementById("new-close")
		var create = document.getElementById("new-create")
		var id = document.getElementById("new-id")
		var model = document.getElementById("new-model")
		var name = document.getElementById("new-name")
		var err = document.getElementById("create-err")

		close.onclick = (event) => this.closeNewDialog(event)
		create.onclick = (event) => this.create(event)

		id.value = this.generateRandomId()
		model.value = ""
		name.value = ""
		err.innerText = ""

		this.loadModels()
	}

	showDeleteDialog() {
		var close = document.getElementById("delete-close")
		var deleteBtn = document.getElementById("delete-delete")
		close.onclick = (event) => this.deleteDialog.close()
		deleteBtn.onclick = (event) => this.deletef()

		var err = document.getElementById("delete-err")
		err.innerText = ""

		var id = this.activeId
		var child = this.state.Children[id]

		var delprompt = document.getElementById("delete-prompt")
		delprompt.innerHTML = "Delete device?<br><br>ID: " + id + "<br>Model: " +
			child.Model + "<br>Name: " + child.Name

		this.deleteDialog.showModal()
	}

	showAboutDialog() {
		var aboutDialog = document.getElementById("about-dialog")
		var close = document.getElementById("about-close")
		close.onclick = (event) => aboutDialog.close()
		aboutDialog.showModal()
	}

	toggleMenu() {
		menu.style.display = (menu.style.display === "block") ? "none" : "block";
	}

	menuItemClick(item) {
		menu.style.display = "none"
		switch (item.textContent) {
		case "Download":
			console.log("Download")
			break;
		case "About":
			this.showAboutDialog()
			break;
		}
	}
}
